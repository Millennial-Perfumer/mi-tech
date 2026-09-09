import { useEffect, useMemo, useState } from "react";
import { Download, FilePlus2, Minus, Plus, Save, X } from "lucide-react";
import {
  apiJson,
  apiRequest,
  arrayFrom,
  formatDate,
  numberValue,
  textValue,
} from "../../lib/http";

export type B2BDocumentKind =
  "invoices" | "proformas" | "credit-notes" | "debit-notes";
export type B2BDocumentMode = "create" | "edit" | "view";
type Row = Record<string, unknown>;

type Props = {
  kind: B2BDocumentKind;
  mode: B2BDocumentMode;
  record?: Row;
  customers: Row[];
  invoices: Row[];
  token: string;
  onUnauthorized: () => void;
  onClose: () => void;
  onSaved: () => void;
  onAction: (action: string, payload?: Record<string, string>) => void;
};

type LineItem = {
  id?: number;
  product_id: string;
  item_details: string;
  sku: string;
  hsn_code: string;
  gst_rate: string;
  quantity: string;
  rate: string;
};

type FormState = {
  id?: number;
  customer_id: string;
  customer_name: string;
  customer_gstin: string;
  customer_email: string;
  customer_phone: string;
  customer_state: string;
  customer_state_code: string;
  customer_address: string;
  customer_shipping_address: string;
  invoice_id: string;
  invoice_number: string;
  invoice_date: string;
  due_date: string;
  note_date: string;
  valid_until: string;
  order_number: string;
  salesperson: string;
  subject: string;
  terms: string;
  reason: string;
  customer_notes: string;
  discount_percent: string;
  transportation_charge: string;
  tds_tcs_type: string;
  tds_tcs_rate: string;
  paid_amount: string;
  payment_method: string;
  items: LineItem[];
};

type PaymentTerm = { id?: number; name: string; days: number };

const today = () => new Date().toISOString().slice(0, 10);

const DEFAULT_PAYMENT_TERMS: PaymentTerm[] = [
  { name: "Net 30", days: 30 },
  { name: "Net 60", days: 60 },
];

const DEFAULT_UPI_ID = "7904769823@hdfc";

const DEFAULT_CUSTOMER_NOTES = `Thanks for your business.

Payment Terms: Full payment is required before the due date mentioned on the invoice.

No Refunds & Returns: Due to the nature of our products, we do not accept returns or provide refunds once the item has been opened or used. If the product remains sealed and unused, you may contact us within 7 days for return eligibility, subject to approval.

Damaged or Incorrect Items: If you receive a damaged or incorrect product, please contact us within 48 hours of delivery with photographic evidence for a replacement or resolution.

Shipping & Delivery: We aim to deliver orders promptly, but delays due to courier services, customs, or unforeseen circumstances are beyond our control. Tracking details will be provided once your order is shipped.

Intellectual Property: All branding, packaging, and product names are trademarks of Millennial Perfumer™ and may not be reproduced without permission.`;

function paymentTermDays(
  value: string,
  terms: PaymentTerm[] = DEFAULT_PAYMENT_TERMS,
) {
  const normalized = value.trim().toLowerCase();
  const known = terms.find((term) => term.name.toLowerCase() === normalized);
  if (known) return known.days;
  if (normalized === "due on receipt") return 0;
  const match = normalized.match(/^net\s+(\d+)$/);
  return match ? Number(match[1]) : undefined;
}

function addDays(value: string, days: number | undefined) {
  if (!value || days === undefined) return "";
  const [year, month, day] = value.split("-").map(Number);
  if (![year, month, day].every(Number.isFinite)) return "";
  return new Date(Date.UTC(year, month - 1, day + days))
    .toISOString()
    .slice(0, 10);
}

function dateInput(value: unknown, fallback = today()) {
  if (!value) return fallback;
  const parsed = new Date(String(value));
  return Number.isNaN(parsed.getTime())
    ? fallback
    : parsed.toISOString().slice(0, 10);
}

function money(value: unknown) {
  return `₹${numberValue(value).toLocaleString("en-IN", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function lineItem(value?: Row): LineItem {
  return {
    id: typeof value?.id === "number" ? value.id : undefined,
    product_id: textValue(value?.product_id, ""),
    item_details: textValue(value?.item_details, ""),
    sku: textValue(value?.sku, ""),
    hsn_code: textValue(value?.hsn_code, "33029019"),
    gst_rate:
      value?.gst_rate === undefined
        ? "18"
        : String(numberValue(value.gst_rate)),
    quantity: value ? String(numberValue(value.quantity)) : "1",
    rate: value ? String(numberValue(value.rate)) : "0",
  };
}

function kindLabel(kind: B2BDocumentKind) {
  return kind === "invoices"
    ? "invoice"
    : kind === "proformas"
      ? "proforma"
      : kind === "credit-notes"
        ? "credit note"
        : "debit note";
}

function collectionLabel(kind: B2BDocumentKind) {
  return kindLabel(kind).replace(/^./, (letter) => letter.toUpperCase());
}

function createForm(kind: B2BDocumentKind, record?: Row): FormState {
  const items = Array.isArray(record?.items)
    ? record.items.map((item) => lineItem(item as Row))
    : [lineItem()];
  const invoiceDate = dateInput(record?.invoice_date);
  const terms = textValue(record?.terms, kind === "invoices" ? "Net 30" : "");
  const customerNotes = record
    ? textValue(record.customer_notes, "")
    : kind === "invoices"
      ? DEFAULT_CUSTOMER_NOTES
      : "";
  return {
    id: typeof record?.id === "number" ? record.id : undefined,
    customer_id: textValue(record?.customer_id, ""),
    customer_name: textValue(record?.customer_name, ""),
    customer_gstin: textValue(record?.customer_gstin, ""),
    customer_email: textValue(record?.customer_email, ""),
    customer_phone: textValue(record?.customer_phone, ""),
    customer_state: textValue(record?.customer_state, ""),
    customer_state_code: textValue(record?.customer_state_code, ""),
    customer_address: textValue(record?.customer_address, ""),
    customer_shipping_address: textValue(record?.customer_shipping_address, ""),
    invoice_id: textValue(record?.invoice_id, ""),
    invoice_number: textValue(record?.invoice_number, ""),
    invoice_date: invoiceDate,
    due_date:
      dateInput(record?.due_date, "") ||
      addDays(invoiceDate, paymentTermDays(terms)),
    note_date: dateInput(record?.note_date),
    valid_until: dateInput(record?.valid_until, ""),
    order_number: textValue(record?.order_number, ""),
    salesperson: textValue(record?.salesperson, ""),
    subject: textValue(record?.subject, ""),
    terms,
    reason: textValue(record?.reason, ""),
    customer_notes: customerNotes,
    discount_percent: String(numberValue(record?.discount_percent)),
    transportation_charge: String(numberValue(record?.transportation_charge)),
    tds_tcs_type: textValue(record?.tds_tcs_type, "NONE"),
    tds_tcs_rate: String(numberValue(record?.tds_tcs_rate)),
    paid_amount: String(numberValue(record?.paid_amount)),
    payment_method: textValue(record?.payment_method, "Bank transfer"),
    items: items.length > 0 ? items : [lineItem()],
  };
}

function customerName(customer: Row) {
  return textValue(
    customer.trade_name || customer.legal_name,
    "Unnamed business",
  );
}

function configMap(value: unknown) {
  const result: Record<string, string> = {};
  for (const config of arrayFrom(value, "configs")) {
    const key = textValue(config.key, "");
    if (key) result[key] = textValue(config.value, "");
  }
  return result;
}

function statusClass(status: string) {
  return status === "CANCELLED"
    ? "danger"
    : ["ISSUED", "PAID", "ACCEPTED"].includes(status)
      ? "success"
      : "neutral";
}

export function B2BDocumentModal({
  kind,
  mode,
  record,
  customers,
  invoices,
  token,
  onUnauthorized,
  onClose,
  onSaved,
  onAction,
}: Props) {
  const [form, setForm] = useState<FormState>(() => createForm(kind, record));
  const [inventoryProducts, setInventoryProducts] = useState<Row[]>([]);
  const [appConfigs, setAppConfigs] = useState<Record<string, string>>({});
  const [paymentTerms, setPaymentTerms] = useState<PaymentTerm[]>(
    DEFAULT_PAYMENT_TERMS,
  );
  const [nextInvoiceNumber, setNextInvoiceNumber] = useState("");
  const [newTermName, setNewTermName] = useState("");
  const [newTermDays, setNewTermDays] = useState("30");
  const [showTermEditor, setShowTermEditor] = useState(false);
  const [isWorking, setIsWorking] = useState(false);
  const [error, setError] = useState("");
  const [printRequest, setPrintRequest] = useState<{
    form: FormState;
    record: Row;
  }>();
  const isReadOnly = mode === "view";
  const label = collectionLabel(kind);

  useEffect(() => {
    setForm(createForm(kind, record));
    setError("");
    setPrintRequest(undefined);
    setNextInvoiceNumber("");
    setShowTermEditor(false);
  }, [kind, record]);

  useEffect(() => {
    let active = true;
    const loadSupportingData = async () => {
      const [termsResult, inventoryResult, configsResult] =
        await Promise.allSettled([
          apiJson<unknown>(token, onUnauthorized, "/api/b2b/payment-terms"),
          apiJson<unknown>(token, onUnauthorized, "/api/inventory"),
          apiJson<unknown>(token, onUnauthorized, "/api/configs"),
        ]);
      if (!active) return;
      if (termsResult.status === "fulfilled") {
        const loadedTerms = arrayFrom(termsResult.value, "terms")
          .map((term) => ({
            id: typeof term.id === "number" ? term.id : undefined,
            name: textValue(term.name, ""),
            days: Math.max(0, numberValue(term.due_days)),
          }))
          .filter((term) => term.name);
        const merged = [...DEFAULT_PAYMENT_TERMS, ...loadedTerms].filter(
          (term, index, all) =>
            all.findIndex(
              (candidate) =>
                candidate.name.toLowerCase() === term.name.toLowerCase(),
            ) === index,
        );
        setPaymentTerms(merged);
      }
      if (inventoryResult.status === "fulfilled")
        setInventoryProducts(arrayFrom(inventoryResult.value, "items"));
      if (configsResult.status === "fulfilled")
        setAppConfigs(configMap(configsResult.value));
    };
    void loadSupportingData();
    return () => {
      active = false;
    };
  }, [onUnauthorized, token]);

  useEffect(() => {
    if (kind !== "invoices" || mode === "view" || !form.invoice_date)
      return undefined;
    let active = true;
    void apiJson<Row>(
      token,
      onUnauthorized,
      `/api/b2b/invoices/next-number?date=${encodeURIComponent(form.invoice_date)}`,
    )
      .then((data) => {
        if (active)
          setNextInvoiceNumber(textValue(data.next_invoice_number, ""));
      })
      .catch(() => {
        if (active) setNextInvoiceNumber("");
      });
    return () => {
      active = false;
    };
  }, [form.invoice_date, kind, mode, onUnauthorized, token]);

  useEffect(() => {
    if (!printRequest) return;
    const previousTitle = document.title;
    const invoiceNumber = textValue(
      printRequest.record.invoice_number,
      "draft",
    );
    document.title = `Invoice-${invoiceNumber}`;
    const finishPrint = () => {
      document.title = previousTitle;
      setPrintRequest(undefined);
      onSaved();
    };
    window.addEventListener("afterprint", finishPrint, { once: true });
    const timer = window.setTimeout(() => window.print(), 0);
    return () => {
      window.clearTimeout(timer);
      window.removeEventListener("afterprint", finishPrint);
      document.title = previousTitle;
    };
  }, [onSaved, printRequest]);

  const sellerGSTIN =
    appConfigs.business_gstin || textValue(record?.seller_gstin, "");
  const sellerStateCode =
    appConfigs.business_gstin?.slice(0, 2) ||
    textValue(record?.seller_state_code, "");
  const sellerName =
    appConfigs.business_name || textValue(record?.seller_name, "Business");
  const sellerAddress =
    [appConfigs.business_address_line1, appConfigs.business_address_line2]
      .filter(Boolean)
      .join(", ") || textValue(record?.seller_address, "");
  const bankName = appConfigs.bank_name || "";
  const bankAccount = appConfigs.bank_account_no || "";
  const bankIfsc = appConfigs.bank_ifsc || "";
  const configuredUpiId = appConfigs.upi_id?.trim() || "";
  const upiId =
    configuredUpiId && configuredUpiId !== "parfumtraders@upi"
      ? configuredUpiId
      : DEFAULT_UPI_ID;

  const totals = useMemo(() => {
    const lines = form.items.map((item) => {
      const amount = numberValue(item.quantity) * numberValue(item.rate);
      const gstRate = numberValue(item.gst_rate);
      return { amount, gstRate, gstAmount: (amount * gstRate) / 100 };
    });
    const subtotal = lines.reduce((sum, item) => sum + item.amount, 0);
    const discount = (subtotal * numberValue(form.discount_percent)) / 100;
    const taxable = Math.max(0, subtotal - discount);
    const discountRatio = subtotal > 0 ? taxable / subtotal : 1;
    const sameState =
      sellerStateCode !== "" &&
      form.customer_state_code !== "" &&
      sellerStateCode === form.customer_state_code;
    let cgst = 0;
    let sgst = 0;
    let igst = 0;
    let cgstRate = 0;
    let sgstRate = 0;
    let igstRate = 0;
    for (const line of lines) {
      const lineTaxable = line.amount * discountRatio;
      if (sameState) {
        cgstRate = line.gstRate / 2;
        sgstRate = line.gstRate / 2;
        cgst += (lineTaxable * cgstRate) / 100;
        sgst += (lineTaxable * sgstRate) / 100;
      } else {
        igstRate = line.gstRate;
        igst += (lineTaxable * igstRate) / 100;
      }
    }
    const transport =
      kind === "invoices" ? numberValue(form.transportation_charge) : 0;
    const transportTaxable = transport / 1.18;
    const transportGST = transport - transportTaxable;
    const tdsTcs =
      form.tds_tcs_type === "NONE"
        ? 0
        : (taxable * numberValue(form.tds_tcs_rate)) / 100;
    const totalTax = cgst + sgst + igst;
    const total =
      taxable +
      totalTax +
      transport +
      (form.tds_tcs_type === "TCS"
        ? tdsTcs
        : form.tds_tcs_type === "TDS"
          ? -tdsTcs
          : 0);
    return {
      lines,
      subtotal,
      discount,
      taxable,
      cgst,
      sgst,
      igst,
      cgstRate,
      sgstRate,
      igstRate,
      totalTax,
      transport,
      transportTaxable,
      transportGST,
      tdsTcs,
      total,
    };
  }, [
    form.customer_state_code,
    form.discount_percent,
    form.items,
    form.tds_tcs_rate,
    form.tds_tcs_type,
    form.transportation_charge,
    kind,
    sellerStateCode,
  ]);

  const updateField = (field: keyof FormState, value: string) =>
    setForm((current) => ({ ...current, [field]: value }));

  const updateInvoiceDate = (invoiceDate: string) =>
    setForm((current) => ({
      ...current,
      invoice_date: invoiceDate,
      due_date: addDays(
        invoiceDate,
        paymentTermDays(current.terms, paymentTerms),
      ),
    }));

  const updatePaymentTerms = (terms: string) =>
    setForm((current) => ({
      ...current,
      terms,
      due_date: addDays(
        current.invoice_date,
        paymentTermDays(terms, paymentTerms),
      ),
    }));

  const selectCustomer = (id: string) => {
    const customer = customers.find((item) => String(item.id) === id);
    if (!customer) {
      updateField("customer_id", "");
      return;
    }
    setForm((current) => ({
      ...current,
      customer_id: id,
      customer_name: customerName(customer),
      customer_gstin: textValue(customer.gstin, ""),
      customer_email: textValue(customer.email, ""),
      customer_phone: textValue(customer.phone, ""),
      customer_state: textValue(customer.state, ""),
      customer_state_code: textValue(
        customer.state_code,
        textValue(customer.gstin, "").slice(0, 2),
      ),
      customer_address: textValue(customer.billing_address, ""),
      customer_shipping_address: textValue(
        customer.shipping_address,
        textValue(customer.billing_address, ""),
      ),
    }));
  };

  const selectInvoice = (id: string) => {
    const invoice = invoices.find((item) => String(item.id) === id);
    if (!invoice) {
      updateField("invoice_id", "");
      return;
    }
    const invoiceItems = Array.isArray(invoice.items)
      ? invoice.items.map((item) => lineItem(item as Row))
      : form.items;
    setForm((current) => ({
      ...current,
      invoice_id: id,
      customer_id: textValue(invoice.customer_id, current.customer_id),
      customer_name: textValue(invoice.customer_name, current.customer_name),
      customer_gstin: textValue(invoice.customer_gstin, current.customer_gstin),
      customer_email: textValue(invoice.customer_email, current.customer_email),
      customer_phone: textValue(invoice.customer_phone, current.customer_phone),
      customer_state: textValue(invoice.customer_state, current.customer_state),
      customer_state_code: textValue(
        invoice.customer_state_code,
        current.customer_state_code,
      ),
      customer_address: textValue(
        invoice.customer_address,
        current.customer_address,
      ),
      customer_shipping_address: textValue(
        invoice.customer_shipping_address,
        current.customer_shipping_address,
      ),
      items: invoiceItems,
    }));
  };

  const updateItem = (index: number, field: keyof LineItem, value: string) =>
    setForm((current) => ({
      ...current,
      items: current.items.map((item, itemIndex) =>
        itemIndex === index ? { ...item, [field]: value } : item,
      ),
    }));

  const updateItemDetails = (index: number, value: string) => {
    const product = inventoryProducts.find(
      (item) =>
        textValue(item.title, "") === value ||
        `${textValue(item.title, "")} (${textValue(item.mi_sku, "")})` ===
          value,
    );
    setForm((current) => ({
      ...current,
      items: current.items.map((item, itemIndex) =>
        itemIndex !== index
          ? item
          : product
            ? {
                ...item,
                product_id: textValue(product.id, ""),
                item_details: textValue(product.title, value),
                sku: textValue(product.mi_sku, item.sku),
                rate: String(numberValue(product.price)),
                hsn_code: textValue(
                  product.hsn_code,
                  item.hsn_code || "33029019",
                ),
              }
            : { ...item, item_details: value },
      ),
    }));
  };

  const saveNewPaymentTerm = async () => {
    const name = newTermName.trim();
    const days = Math.max(0, Math.floor(numberValue(newTermDays)));
    if (!name) {
      setError("Enter a payment term name");
      return;
    }
    setIsWorking(true);
    setError("");
    try {
      const response = await apiJson<Row>(
        token,
        onUnauthorized,
        "/api/b2b/payment-terms",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ name, due_days: days }),
        },
      );
      const savedTerm: PaymentTerm = {
        id: typeof response.id === "number" ? response.id : undefined,
        name: textValue(response.name, name),
        days: numberValue(response.due_days),
      };
      setPaymentTerms((current) => [
        ...current.filter(
          (term) => term.name.toLowerCase() !== savedTerm.name.toLowerCase(),
        ),
        savedTerm,
      ]);
      updatePaymentTerms(savedTerm.name);
      setNewTermName("");
      setShowTermEditor(false);
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Unable to save payment term",
      );
    } finally {
      setIsWorking(false);
    }
  };

  const save = async (
    event: { preventDefault: () => void },
    issueAfterSave = false,
    downloadAfterSave = false,
  ) => {
    event.preventDefault();
    if (!form.customer_id || !form.customer_gstin) {
      setError("Select a business customer before saving the document");
      return;
    }
    setIsWorking(true);
    setError("");
    try {
      const payload: Record<string, unknown> = {
        id: form.id,
        customer_id: form.customer_id ? Number(form.customer_id) : null,
        customer_name: form.customer_name,
        customer_gstin: form.customer_gstin.toUpperCase(),
        customer_email: form.customer_email || null,
        customer_phone: form.customer_phone || null,
        customer_state: form.customer_state,
        customer_state_code: form.customer_state_code,
        customer_address: form.customer_address,
        customer_shipping_address: form.customer_shipping_address,
        invoice_id: form.invoice_id ? Number(form.invoice_id) : null,
        invoice_number: form.invoice_number || null,
        order_number: form.order_number || null,
        salesperson: form.salesperson || null,
        subject: form.subject || null,
        terms: form.terms || null,
        reason: form.reason || null,
        customer_notes: form.customer_notes || null,
        items: form.items.map((item) => ({
          id: item.id,
          product_id: item.product_id ? Number(item.product_id) : null,
          item_details: item.item_details,
          sku: item.sku || null,
          hsn_code: item.hsn_code || null,
          gst_rate: numberValue(item.gst_rate),
          quantity: numberValue(item.quantity),
          rate: numberValue(item.rate),
          amount: numberValue(item.quantity) * numberValue(item.rate),
        })),
        discount_percent: numberValue(form.discount_percent),
        transportation_charge: numberValue(form.transportation_charge),
        tds_tcs_type: form.tds_tcs_type,
        tds_tcs_rate: numberValue(form.tds_tcs_rate),
        paid_amount: numberValue(form.paid_amount),
      };
      delete payload.id;
      if (form.id) payload.id = form.id;
      if (kind === "invoices") {
        payload.invoice_date = `${form.invoice_date}T00:00:00Z`;
        payload.due_date = form.due_date ? `${form.due_date}T00:00:00Z` : null;
      } else if (kind === "proformas") {
        payload.note_date = `${form.note_date}T00:00:00Z`;
        payload.valid_until = form.valid_until
          ? `${form.valid_until}T00:00:00Z`
          : null;
      } else {
        payload.note_date = `${form.note_date}T00:00:00Z`;
      }
      const response = await apiRequest(
        token,
        onUnauthorized,
        `/api/b2b/${kind}`,
        {
          method: form.id ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        },
      );
      let saved: Row = {};
      try {
        saved = (await response.json()) as Row;
      } catch {
        /* Some deployments return an empty body. */
      }
      let savedRecord: Row = { ...form, ...saved };
      if (issueAfterSave) {
        const savedId =
          typeof saved.id === "number"
            ? String(saved.id)
            : textValue(saved.id, form.id ? String(form.id) : "");
        const issueResponse = await apiRequest(
          token,
          onUnauthorized,
          `/api/b2b/${kind}/issue?id=${savedId}`,
          { method: "POST" },
        );
        try {
          savedRecord = {
            ...savedRecord,
            ...((await issueResponse.json()) as Row),
          };
        } catch {
          /* The saved document is still printable. */
        }
      }
      if (downloadAfterSave && kind === "invoices") {
        setPrintRequest({
          form: {
            ...form,
            id: typeof saved.id === "number" ? saved.id : form.id,
          },
          record: savedRecord,
        });
      } else {
        onSaved();
      }
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : `Unable to save ${kindLabel(kind)}`,
      );
    } finally {
      setIsWorking(false);
    }
  };

  const title =
    mode === "create"
      ? `New ${kindLabel(kind)}`
      : mode === "edit"
        ? `Edit ${kindLabel(kind)}`
        : `${label} details`;
  const status = textValue(record?.status, "DRAFT").toUpperCase();
  const canIssue = mode === "view" && status === "DRAFT";
  const printableForm = printRequest?.form || form;
  const printableRecord = printRequest?.record || record;
  const printableTotal =
    printableRecord?.total_price !== undefined
      ? printableRecord.total_price
      : totals.total;
  const printableInvoiceNumber = textValue(
    printableRecord?.invoice_number,
    printableForm.invoice_number || "DRAFT",
  );
  const printableStatus = textValue(
    printableRecord?.status,
    printRequest ? "ISSUED" : status,
  ).toUpperCase();
  const printableQrAmount = numberValue(
    printableRecord?.balance_amount ?? printableTotal,
  );

  const downloadInvoice = () => {
    const previousTitle = document.title;
    document.title = `Invoice-${textValue(record?.invoice_number, record?.id ? String(record.id) : "draft")}`;
    const restoreTitle = () => {
      document.title = previousTitle;
      window.removeEventListener("afterprint", restoreTitle);
    };
    window.addEventListener("afterprint", restoreTitle, { once: true });
    window.print();
  };

  const renderCustomerSummary = () => (
    <div className="b2b-client-card">
      <div className="b2b-client-card-heading">
        <strong>{form.customer_name || "No customer selected"}</strong>
        {form.customer_gstin && (
          <span className="status-pill status-pill-success">GST active</span>
        )}
      </div>
      {form.customer_id ? (
        <>
          <div className="b2b-address-grid">
            <div>
              <span>Billing address</span>
              <p>{form.customer_address || "Not supplied"}</p>
            </div>
            <div>
              <span>Shipping address</span>
              <p>
                {form.customer_shipping_address ||
                  form.customer_address ||
                  "Not supplied"}
              </p>
            </div>
          </div>
          <div className="b2b-client-meta">
            <span>
              <strong>GSTIN:</strong> {form.customer_gstin}
            </span>
            <span>
              <strong>State:</strong> {form.customer_state || "—"} (
              {form.customer_state_code || "—"})
            </span>
            <span>
              <strong>Contact:</strong>{" "}
              {form.customer_email || form.customer_phone || "—"}
            </span>
          </div>
        </>
      ) : (
        <p className="b2b-empty-help">
          Choose a business customer to load GST, address, and contact details.
        </p>
      )}
    </div>
  );

  const renderSummary = (source: Row | undefined = record) => {
    const value = (key: string, fallback: unknown) =>
      source?.[key] === undefined || source?.[key] === null
        ? fallback
        : source[key];
    const subtotal = value("subtotal_price", totals.subtotal);
    const discount = value("discount_amount", totals.discount);
    const cgst = value("cgst_amount", totals.cgst);
    const sgst = value("sgst_amount", totals.sgst);
    const igst = value("igst_amount", totals.igst);
    const transport = value("transportation_charge", totals.transport);
    const adjustment = value("tds_tcs_amount", totals.tdsTcs);
    const total = value("total_price", totals.total);
    const adjustmentType = textValue(
      value("tds_tcs_type", form.tds_tcs_type),
      "NONE",
    );
    return (
      <div className="b2b-breakdown-card">
        <div className="b2b-summary-row">
          <span>Subtotal</span>
          <strong>{money(subtotal)}</strong>
        </div>
        {numberValue(discount) > 0 && (
          <div className="b2b-summary-row">
            <span>
              Discount (
              {textValue(value("discount_percent", form.discount_percent), "0")}
              %)
            </span>
            <span className="b2b-negative">− {money(discount)}</span>
          </div>
        )}
        {numberValue(cgst) > 0 && (
          <div className="b2b-summary-row">
            <span>
              CGST ({numberValue(value("cgst_rate", totals.cgstRate))}%)
            </span>
            <span>{money(cgst)}</span>
          </div>
        )}
        {numberValue(sgst) > 0 && (
          <div className="b2b-summary-row">
            <span>
              SGST ({numberValue(value("sgst_rate", totals.sgstRate))}%)
            </span>
            <span>{money(sgst)}</span>
          </div>
        )}
        {numberValue(igst) > 0 && (
          <div className="b2b-summary-row">
            <span>
              IGST ({numberValue(value("igst_rate", totals.igstRate))}%)
            </span>
            <span>{money(igst)}</span>
          </div>
        )}
        {numberValue(transport) > 0 && (
          <div className="b2b-summary-row b2b-summary-row-stack">
            <span>Transportation</span>
            <span>{money(transport)}</span>
            <small>
              Taxable {money(numberValue(transport) / 1.18)} · GST{" "}
              {money(numberValue(transport) - numberValue(transport) / 1.18)}
            </small>
          </div>
        )}
        {adjustmentType !== "NONE" && numberValue(adjustment) > 0 && (
          <div className="b2b-summary-row b2b-adjustment-row">
            <span>
              {adjustmentType} (
              {textValue(value("tds_tcs_rate", form.tds_tcs_rate), "0")}%)
            </span>
            <span>
              {adjustmentType === "TDS" ? "−" : "+"} {money(adjustment)}
            </span>
          </div>
        )}
        <div className="b2b-summary-total">
          <span>Total amount</span>
          <strong>{money(total)}</strong>
        </div>
      </div>
    );
  };

  return (
    <div
      className="modal-scrim"
      role="presentation"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose();
      }}
    >
      <section
        className={`modal-card b2b-document-modal ${isReadOnly ? "b2b-document-view" : ""}`}
        role="dialog"
        aria-modal="true"
        aria-labelledby="b2b-document-modal-heading"
      >
        <div className="modal-heading">
          <div>
            <p className="eyebrow">B2B billing / {kindLabel(kind)}</p>
            <h2 id="b2b-document-modal-heading">{title}</h2>
            {isReadOnly && (
              <span
                className={`status-pill status-pill-${statusClass(status)}`}
              >
                {status}
              </span>
            )}
          </div>
          <button
            className="icon-button"
            type="button"
            aria-label="Close"
            onClick={onClose}
          >
            <X size={19} aria-hidden="true" />
          </button>
        </div>
        {error && (
          <div className="dashboard-error" role="alert">
            <span>{error}</span>
          </div>
        )}
        {isReadOnly ? (
          <>
            <div className="b2b-document-summary">
              <div>
                <span className="metric-label">Customer</span>
                <strong>{form.customer_name || "—"}</strong>
                <small>{form.customer_gstin || "GSTIN not supplied"}</small>
              </div>
              <div>
                <span className="metric-label">Document date</span>
                <strong>
                  {formatDate(
                    kind === "invoices" ? form.invoice_date : form.note_date,
                  )}
                </strong>
                <small>
                  {record?.due_date
                    ? `Due ${formatDate(record.due_date)}`
                    : record?.valid_until
                      ? `Valid until ${formatDate(record.valid_until)}`
                      : "No secondary date"}
                </small>
              </div>
              <div>
                <span className="metric-label">Total</span>
                <strong>{money(record?.total_price ?? totals.total)}</strong>
                <small>
                  {kind === "invoices"
                    ? `${textValue(record?.payment_status, "UNPAID")} · ${money(record?.balance_amount)}`
                    : "Tax-inclusive total"}
                </small>
              </div>
            </div>
            <div className="b2b-document-address">
              <div>
                <span className="metric-label">Bill to</span>
                <p>{form.customer_address || "No billing address"}</p>
                <span className="metric-label">Ship to</span>
                <p>
                  {form.customer_shipping_address ||
                    form.customer_address ||
                    "No shipping address"}
                </p>
              </div>
              <div>
                <span className="metric-label">Invoice metadata</span>
                <p className="mono-text">
                  {textValue(
                    record?.invoice_number ||
                      record?.proforma_number ||
                      record?.credit_note_number ||
                      record?.debit_note_number,
                    `Draft ${textValue(record?.id)}`,
                  )}
                </p>
                <small>Terms: {form.terms || "—"}</small>
                <small>PO / reference: {form.order_number || "—"}</small>
                <small>Salesperson: {form.salesperson || "—"}</small>
                <small>Subject: {form.subject || "—"}</small>
              </div>
            </div>
            <div className="b2b-items-table">
              <div className="b2b-items-table-header">
                <span>Item / HSN</span>
                <span>Qty</span>
                <span>Rate</span>
                <span>GST</span>
                <span>Amount</span>
              </div>
              {form.items.map((item, index) => (
                <div
                  className="b2b-items-table-row"
                  key={`${item.id || "item"}-${index}`}
                >
                  <span>
                    <strong>{item.item_details || "Unnamed item"}</strong>
                    <small>
                      {item.sku || "SKU —"} · HSN {item.hsn_code || "—"}
                    </small>
                  </span>
                  <span>{item.quantity}</span>
                  <span>{money(item.rate)}</span>
                  <span>{item.gst_rate}%</span>
                  <span>
                    {money(numberValue(item.quantity) * numberValue(item.rate))}
                  </span>
                </div>
              ))}
            </div>
            <div className="b2b-view-lower-grid">
              <div>
                {form.customer_notes && (
                  <div className="b2b-notes-card">
                    <span className="metric-label">
                      Customer notes / terms &amp; conditions
                    </span>
                    <p>{form.customer_notes}</p>
                  </div>
                )}
                {(bankName || bankAccount || bankIfsc || upiId) && (
                  <div className="b2b-payment-details">
                    <div className="b2b-payment-details-heading">
                      <span className="metric-label">Payment details</span>
                      {upiId && (
                        <span className="status-pill status-pill-success">
                          UPI active
                        </span>
                      )}
                    </div>
                    {bankName && (
                      <p>
                        <span>Bank</span>
                        <strong>{bankName}</strong>
                      </p>
                    )}
                    {bankAccount && (
                      <p>
                        <span>Account number</span>
                        <strong>{bankAccount}</strong>
                      </p>
                    )}
                    {bankIfsc && (
                      <p>
                        <span>IFSC</span>
                        <strong>{bankIfsc}</strong>
                      </p>
                    )}
                    {upiId && (
                      <p>
                        <span>UPI ID</span>
                        <strong>{upiId}</strong>
                      </p>
                    )}
                  </div>
                )}
              </div>
              {renderSummary()}
            </div>
            {kind === "invoices" && status === "ISSUED" && (
              <div className="b2b-payment-box">
                <div>
                  <span className="metric-label">Record payment</span>
                  <p>
                    Update the amount collected without changing the issued
                    document.
                  </p>
                </div>
                <div className="form-grid-two">
                  <label className="form-field">
                    <span>Paid amount</span>
                    <input
                      type="number"
                      min="0"
                      step="0.01"
                      value={form.paid_amount}
                      onChange={(event) =>
                        updateField("paid_amount", event.target.value)
                      }
                    />
                  </label>
                  <label className="form-field">
                    <span>Payment method</span>
                    <select
                      value={form.payment_method}
                      onChange={(event) =>
                        updateField("payment_method", event.target.value)
                      }
                    >
                      <option>Bank transfer</option>
                      <option>UPI</option>
                      <option>Cash</option>
                      <option>Cheque</option>
                      <option>Other</option>
                    </select>
                  </label>
                </div>
                <button
                  className="secondary-button"
                  type="button"
                  disabled={isWorking}
                  onClick={() =>
                    onAction("payment", {
                      paid_amount: form.paid_amount,
                      payment_method: form.payment_method,
                    })
                  }
                >
                  <Save size={14} aria-hidden="true" /> Save payment
                </button>
              </div>
            )}
            <div className="modal-actions b2b-document-actions no-print">
              <div className="table-action-group">
                {status === "DRAFT" && (
                  <button
                    className="table-link-button"
                    type="button"
                    onClick={() => onAction("edit")}
                  >
                    Edit
                  </button>
                )}
                {status === "DRAFT" && (
                  <button
                    className="table-link-button danger-link"
                    type="button"
                    onClick={() => onAction("delete")}
                  >
                    Delete
                  </button>
                )}
                {canIssue && (
                  <button
                    className="primary-button"
                    type="button"
                    disabled={isWorking}
                    onClick={() => onAction("issue")}
                  >
                    Issue
                  </button>
                )}
                {kind === "invoices" && status === "ISSUED" && (
                  <>
                    {record?.inventory_deducted ? (
                      <button
                        className="table-link-button"
                        type="button"
                        onClick={() => onAction("revert-inventory")}
                      >
                        Revert stock
                      </button>
                    ) : (
                      <button
                        className="table-link-button"
                        type="button"
                        onClick={() => onAction("deduct-inventory")}
                      >
                        Deduct stock
                      </button>
                    )}
                    <button
                      className="table-link-button danger-link"
                      type="button"
                      onClick={() => onAction("cancel")}
                    >
                      Cancel invoice
                    </button>
                  </>
                )}
                {kind === "proformas" && status === "SENT" && (
                  <>
                    <button
                      className="table-link-button"
                      type="button"
                      onClick={() => onAction("accept")}
                    >
                      Accept
                    </button>
                    <button
                      className="table-link-button"
                      type="button"
                      onClick={() => onAction("reject")}
                    >
                      Reject
                    </button>
                  </>
                )}
                {kind === "proformas" && status === "ACCEPTED" && (
                  <button
                    className="primary-button"
                    type="button"
                    onClick={() => onAction("convert")}
                  >
                    Convert to invoice
                  </button>
                )}
                {kind === "proformas" &&
                  ["SENT", "ACCEPTED"].includes(status) && (
                    <button
                      className="table-link-button danger-link"
                      type="button"
                      onClick={() => onAction("cancel")}
                    >
                      Cancel
                    </button>
                  )}
                {["credit-notes", "debit-notes"].includes(kind) &&
                  status === "ISSUED" && (
                    <button
                      className="table-link-button danger-link"
                      type="button"
                      onClick={() => onAction("cancel")}
                    >
                      Cancel
                    </button>
                  )}
              </div>
              {kind === "invoices" && (
                <button
                  className="secondary-button"
                  type="button"
                  onClick={downloadInvoice}
                >
                  <Download size={14} aria-hidden="true" /> Download PDF
                </button>
              )}
              <button
                className="secondary-button"
                type="button"
                onClick={onClose}
              >
                Close
              </button>
            </div>
          </>
        ) : (
          <form onSubmit={(event) => void save(event)}>
            <section className="b2b-form-section b2b-client-section">
              <div className="b2b-form-section-heading">
                <div>
                  <p className="eyebrow">Client information</p>
                  <h3>Select business customer</h3>
                </div>
              </div>
              <label className="form-field">
                <span>
                  Customer <small>(required)</small>
                </span>
                <select
                  required
                  value={form.customer_id}
                  onChange={(event) => selectCustomer(event.target.value)}
                >
                  <option value="">Select business customer</option>
                  {customers.map((customer) => (
                    <option
                      key={String(customer.id)}
                      value={String(customer.id)}
                    >
                      {customerName(customer)} ·{" "}
                      {textValue(customer.gstin, "GSTIN missing")}
                    </option>
                  ))}
                </select>
              </label>
              {form.customer_id && (
                <div className="form-grid-two">
                  <label className="form-field">
                    <span>Billing address</span>
                    <textarea
                      required
                      rows={3}
                      value={form.customer_address}
                      onChange={(event) =>
                        updateField("customer_address", event.target.value)
                      }
                    />
                  </label>
                  <label className="form-field">
                    <span>Shipping address</span>
                    <textarea
                      rows={3}
                      value={form.customer_shipping_address}
                      onChange={(event) =>
                        updateField(
                          "customer_shipping_address",
                          event.target.value,
                        )
                      }
                    />
                  </label>
                </div>
              )}
              {renderCustomerSummary()}
            </section>
            <section className="b2b-form-section">
              <div className="b2b-form-section-heading">
                <div>
                  <p className="eyebrow">Billing details</p>
                  <h3>Dates, reference, and terms</h3>
                </div>
              </div>
              <div className="form-grid-two">
                <label className="form-field">
                  <span>
                    {kind === "invoices" ? "Invoice date" : "Note date"}{" "}
                    <small>(required)</small>
                  </span>
                  <input
                    required
                    type="date"
                    value={
                      kind === "invoices" ? form.invoice_date : form.note_date
                    }
                    onChange={(event) =>
                      kind === "invoices"
                        ? updateInvoiceDate(event.target.value)
                        : updateField("note_date", event.target.value)
                    }
                  />
                </label>
                {kind === "invoices" ? (
                  <label className="form-field">
                    <span>Due date</span>
                    <input
                      type="date"
                      value={form.due_date}
                      onChange={(event) =>
                        updateField("due_date", event.target.value)
                      }
                    />
                  </label>
                ) : kind === "proformas" ? (
                  <label className="form-field">
                    <span>Valid until</span>
                    <input
                      type="date"
                      value={form.valid_until}
                      onChange={(event) =>
                        updateField("valid_until", event.target.value)
                      }
                    />
                  </label>
                ) : (
                  <label className="form-field">
                    <span>Linked invoice</span>
                    <select
                      value={form.invoice_id}
                      onChange={(event) => selectInvoice(event.target.value)}
                    >
                      <option value="">Select invoice (optional)</option>
                      {invoices.map((invoice) => (
                        <option
                          key={String(invoice.id)}
                          value={String(invoice.id)}
                        >
                          {textValue(
                            invoice.invoice_number,
                            `Invoice ${invoice.id}`,
                          )}{" "}
                          · {textValue(invoice.customer_name, "Customer")}
                        </option>
                      ))}
                    </select>
                  </label>
                )}
              </div>
              {kind === "invoices" && (
                <div className="form-grid-three">
                  <label className="form-field">
                    <span>Invoice number</span>
                    <div className="b2b-readonly-field">
                      <strong>
                        {form.invoice_number ||
                          nextInvoiceNumber ||
                          "Auto-generated on issue"}
                      </strong>
                      <small>read-only</small>
                    </div>
                  </label>
                  <label className="form-field">
                    <span>Salesperson</span>
                    <input
                      value={form.salesperson}
                      onChange={(event) =>
                        updateField("salesperson", event.target.value)
                      }
                      placeholder="e.g. John Doe"
                    />
                  </label>
                  <label className="form-field">
                    <span>Payment terms</span>
                    <select
                      value={form.terms}
                      onChange={(event) =>
                        event.target.value === "__CREATE_NEW__"
                          ? setShowTermEditor(true)
                          : updatePaymentTerms(event.target.value)
                      }
                      aria-describedby="payment-terms-help"
                    >
                      <option value="">Select terms</option>
                      {paymentTerms.map((term) => (
                        <option
                          key={`${term.name}-${term.id || "default"}`}
                          value={term.name}
                        >
                          {term.name}
                        </option>
                      ))}
                      <option value="__CREATE_NEW__">+ New payment term</option>
                    </select>
                    <small
                      id="payment-terms-help"
                      className="payment-terms-help"
                    >
                      Selecting Net 30 or Net 60 updates the due date
                      automatically.
                    </small>
                  </label>
                </div>
              )}
              {showTermEditor && kind === "invoices" && (
                <div className="b2b-payment-term-editor">
                  <label className="form-field">
                    <span>Term name</span>
                    <input
                      value={newTermName}
                      onChange={(event) => setNewTermName(event.target.value)}
                      placeholder="e.g. Net 90"
                    />
                  </label>
                  <label className="form-field">
                    <span>Due in days</span>
                    <input
                      type="number"
                      min="0"
                      step="1"
                      value={newTermDays}
                      onChange={(event) => setNewTermDays(event.target.value)}
                    />
                  </label>
                  <button
                    className="secondary-button"
                    type="button"
                    disabled={isWorking}
                    onClick={() => void saveNewPaymentTerm()}
                  >
                    Save term
                  </button>
                  <button
                    className="table-link-button"
                    type="button"
                    onClick={() => setShowTermEditor(false)}
                  >
                    Cancel
                  </button>
                </div>
              )}
              {kind === "invoices" && (
                <div className="form-grid-two">
                  <label className="form-field">
                    <span>Order / PO reference</span>
                    <input
                      value={form.order_number}
                      onChange={(event) =>
                        updateField("order_number", event.target.value)
                      }
                      placeholder="Optional order or PO number"
                    />
                  </label>
                  <label className="form-field">
                    <span>Subject</span>
                    <input
                      value={form.subject}
                      onChange={(event) =>
                        updateField("subject", event.target.value)
                      }
                      placeholder="e.g. Supply of fragrance ingredients"
                    />
                  </label>
                </div>
              )}
              {["credit-notes", "debit-notes"].includes(kind) && (
                <label className="form-field">
                  <span>Reason</span>
                  <input
                    required
                    value={form.reason}
                    onChange={(event) =>
                      updateField("reason", event.target.value)
                    }
                    placeholder="Why is this adjustment being raised?"
                  />
                </label>
              )}
            </section>
            <section className="b2b-form-section">
              <div className="b2b-form-section-heading">
                <div>
                  <p className="eyebrow">Items table</p>
                  <h3>Products and services</h3>
                </div>
                <button
                  className="secondary-button"
                  type="button"
                  onClick={() =>
                    setForm((current) => ({
                      ...current,
                      items: [...current.items, lineItem()],
                    }))
                  }
                >
                  <Plus size={14} aria-hidden="true" /> Add row
                </button>
              </div>
              <div className="b2b-items-edit-wrap">
                <div className="b2b-items-edit-table">
                  <div className="b2b-items-edit-header">
                    <span>Item details</span>
                    <span>HSN</span>
                    <span>GST %</span>
                    <span>Quantity</span>
                    <span>Rate (₹)</span>
                    <span>GST (₹)</span>
                    <span>Amount</span>
                    <span />
                  </div>
                  {form.items.map((item, index) => (
                    <div
                      className="b2b-items-edit-row"
                      key={`${item.id || "new"}-${index}`}
                    >
                      <label className="form-field">
                        <span className="mobile-only-label">Item details</span>
                        <input
                          required
                          list={`b2b-products-${index}`}
                          value={item.item_details}
                          onChange={(event) =>
                            updateItemDetails(index, event.target.value)
                          }
                          placeholder="Search warehouse products or type custom details"
                        />
                        <datalist id={`b2b-products-${index}`}>
                          {inventoryProducts.map((product) => (
                            <option
                              key={String(product.id)}
                              value={
                                textValue(product.title, "") +
                                " (" +
                                textValue(product.mi_sku, "") +
                                ")"
                              }
                            />
                          ))}
                        </datalist>
                        <small>
                          {item.sku ||
                            (item.product_id
                              ? "Product selected"
                              : "Custom line item")}
                        </small>
                      </label>
                      <label className="form-field">
                        <span className="mobile-only-label">HSN</span>
                        <input
                          maxLength={8}
                          value={item.hsn_code}
                          onChange={(event) =>
                            updateItem(index, "hsn_code", event.target.value)
                          }
                          placeholder="33029019"
                        />
                      </label>
                      <label className="form-field">
                        <span className="mobile-only-label">GST %</span>
                        <select
                          value={item.gst_rate}
                          onChange={(event) =>
                            updateItem(index, "gst_rate", event.target.value)
                          }
                        >
                          <option value="0">0%</option>
                          <option value="5">5%</option>
                          <option value="12">12%</option>
                          <option value="18">18%</option>
                          <option value="28">28%</option>
                        </select>
                      </label>
                      <label className="form-field">
                        <span className="mobile-only-label">Quantity</span>
                        <input
                          required
                          type="number"
                          min="0.01"
                          step="0.01"
                          value={item.quantity}
                          onChange={(event) =>
                            updateItem(index, "quantity", event.target.value)
                          }
                        />
                        <small>
                          {item.product_id
                            ? `Stock: ${textValue(inventoryProducts.find((product) => String(product.id) === item.product_id)?.current_stock, "0")}`
                            : " "}
                        </small>
                      </label>
                      <label className="form-field">
                        <span className="mobile-only-label">Rate</span>
                        <input
                          required
                          type="number"
                          min="0"
                          step="0.01"
                          value={item.rate}
                          onChange={(event) =>
                            updateItem(index, "rate", event.target.value)
                          }
                        />
                      </label>
                      <span className="b2b-line-total">
                        {money(totals.lines[index]?.gstAmount)}
                      </span>
                      <span className="b2b-line-total">
                        {money(totals.lines[index]?.amount)}
                      </span>
                      <button
                        className="icon-button"
                        type="button"
                        aria-label={`Remove item ${index + 1}`}
                        disabled={form.items.length === 1}
                        onClick={() =>
                          setForm((current) => ({
                            ...current,
                            items: current.items.filter(
                              (_, itemIndex) => itemIndex !== index,
                            ),
                          }))
                        }
                      >
                        <Minus size={15} aria-hidden="true" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </section>
            <div className="b2b-calculation-grid">
              <div className="b2b-calculation-left">
                <label className="form-field">
                  <span>Customer notes / terms &amp; conditions</span>
                  <textarea
                    rows={12}
                    value={form.customer_notes}
                    onChange={(event) =>
                      updateField("customer_notes", event.target.value)
                    }
                    placeholder="Add payment terms, return policy, shipping notes, or other customer instructions."
                  />
                </label>
                <section className="b2b-tax-adjustments">
                  <p className="eyebrow">Tax adjustments</p>
                  <div className="form-grid-two">
                    <label className="form-field">
                      <span>TDS / TCS type</span>
                      <select
                        value={form.tds_tcs_type}
                        onChange={(event) =>
                          updateField("tds_tcs_type", event.target.value)
                        }
                      >
                        <option value="NONE">None</option>
                        <option value="TDS">TDS</option>
                        <option value="TCS">TCS</option>
                      </select>
                    </label>
                    {form.tds_tcs_type !== "NONE" && (
                      <label className="form-field">
                        <span>Rate (%)</span>
                        <input
                          type="number"
                          min="0"
                          step="0.01"
                          value={form.tds_tcs_rate}
                          onChange={(event) =>
                            updateField("tds_tcs_rate", event.target.value)
                          }
                        />
                      </label>
                    )}
                  </div>
                </section>
                {(bankName || bankAccount || bankIfsc || upiId) && (
                  <div className="b2b-payment-details">
                    <div className="b2b-payment-details-heading">
                      <span className="metric-label">Payment details on invoice</span>
                      {upiId && <span className="status-pill status-pill-success">UPI active</span>}
                    </div>
                    {bankName && <p><span>Bank</span><strong>{bankName}</strong></p>}
                    {bankAccount && <p><span>Account number</span><strong>{bankAccount}</strong></p>}
                    {bankIfsc && <p><span>IFSC</span><strong>{bankIfsc}</strong></p>}
                    {upiId && <p><span>UPI ID</span><strong>{upiId}</strong></p>}
                  </div>
                )}
              </div>
              <div>
                {
                  <div className="b2b-edit-breakdown">
                    <div className="b2b-summary-row">
                      <span>Subtotal</span>
                      <strong>{money(totals.subtotal)}</strong>
                    </div>
                    <div className="b2b-summary-row">
                      <label htmlFor="b2b-discount">Discount (%)</label>
                      <input
                        id="b2b-discount"
                        type="number"
                        min="0"
                        max="100"
                        step="0.01"
                        value={form.discount_percent}
                        onChange={(event) =>
                          updateField("discount_percent", event.target.value)
                        }
                      />
                    </div>
                    {totals.discount > 0 && (
                      <div className="b2b-summary-row">
                        <span>Discount amount</span>
                        <span className="b2b-negative">
                          − {money(totals.discount)}
                        </span>
                      </div>
                    )}
                    {totals.cgst > 0 && (
                      <div className="b2b-summary-row">
                        <span>CGST ({totals.cgstRate}%)</span>
                        <span>{money(totals.cgst)}</span>
                      </div>
                    )}
                    {totals.sgst > 0 && (
                      <div className="b2b-summary-row">
                        <span>SGST ({totals.sgstRate}%)</span>
                        <span>{money(totals.sgst)}</span>
                      </div>
                    )}
                    {totals.igst > 0 && (
                      <div className="b2b-summary-row">
                        <span>IGST ({totals.igstRate}%)</span>
                        <span>{money(totals.igst)}</span>
                      </div>
                    )}
                    {kind === "invoices" && (
                      <div className="b2b-summary-row b2b-summary-row-stack">
                        <label htmlFor="b2b-transport">Transportation</label>
                        <input
                          id="b2b-transport"
                          type="number"
                          min="0"
                          step="0.01"
                          value={form.transportation_charge}
                          onChange={(event) =>
                            updateField(
                              "transportation_charge",
                              event.target.value,
                            )
                          }
                        />
                        <small>
                          Taxable {money(totals.transportTaxable)} · GST{" "}
                          {money(totals.transportGST)}
                        </small>
                      </div>
                    )}
                    {form.tds_tcs_type !== "NONE" && totals.tdsTcs > 0 && (
                      <div className="b2b-summary-row b2b-adjustment-row">
                        <span>
                          {form.tds_tcs_type} ({form.tds_tcs_rate}%)
                        </span>
                        <span>
                          {form.tds_tcs_type === "TDS" ? "−" : "+"}{" "}
                          {money(totals.tdsTcs)}
                        </span>
                      </div>
                    )}
                    <div className="b2b-summary-total">
                      <span>Total amount</span>
                      <strong>{money(totals.total)}</strong>
                    </div>
                    <small className="b2b-total-help">
                      GST is recalculated by the backend when the document is
                      saved.
                    </small>
                  </div>
                }
              </div>
            </div>
            <div className="modal-actions b2b-edit-actions no-print">
              <button
                className="secondary-button"
                type="button"
                onClick={onClose}
              >
                Cancel
              </button>
              {kind === "invoices" && (
                <button
                  className="secondary-button"
                  type="button"
                  disabled={isWorking}
                  onClick={(event) => void save(event, false, true)}
                >
                  <Download size={14} aria-hidden="true" /> Save &amp; download PDF
                </button>
              )}
              <button
                className="secondary-button"
                type="submit"
                disabled={isWorking}
              >
                <Save size={14} aria-hidden="true" />{" "}
                {isWorking ? "Saving…" : "Save as draft"}
              </button>
              {kind === "invoices" && (
                <button
                  className="primary-button"
                  type="button"
                  disabled={isWorking}
                  onClick={(event) => void save(event, true)}
                >
                  <FilePlus2 size={14} aria-hidden="true" /> Save &amp; issue
                  bill
                </button>
              )}
            </div>
          </form>
        )}
        {kind === "invoices" && (isReadOnly || printRequest) && (
          <article
            className="invoice-print-preview invoice-pdf-sheet"
            aria-label="Printable invoice"
          >
            <header className="invoice-print-header">
              <div className="invoice-print-brand">
                <p className="invoice-print-kicker">B2B tax invoice</p>
                <h1>Tax invoice</h1>
                <strong>{sellerName}</strong>
                <small>{sellerAddress || "Business address not supplied"}</small>
                {sellerGSTIN && <small>GSTIN: {sellerGSTIN}</small>}
              </div>
              <div className="invoice-print-document-card">
                <span className="invoice-print-status">{printableStatus}</span>
                <div className="invoice-print-document-row">
                  <span>Invoice no.</span>
                  <strong>{printableInvoiceNumber}</strong>
                </div>
                <div className="invoice-print-document-row">
                  <span>Invoice date</span>
                  <strong>{formatDate(printableForm.invoice_date)}</strong>
                </div>
                <div className="invoice-print-document-row">
                  <span>Due date</span>
                  <strong>
                    {printableForm.due_date
                      ? formatDate(printableForm.due_date)
                      : "—"}
                  </strong>
                </div>
              </div>
            </header>
            <div className="invoice-print-meta">
              <div className="invoice-print-party-card">
                <span className="invoice-print-section-label">Bill to</span>
                <strong>
                  {printableForm.customer_name || "Business customer"}
                </strong>
                <small>
                  {printableForm.customer_gstin || "GSTIN not supplied"}
                </small>
                <small>
                  {printableForm.customer_address || "No billing address"}
                </small>
              </div>
              <div className="invoice-print-party-card">
                <span className="invoice-print-section-label">Ship to</span>
                <small>
                  {printableForm.customer_shipping_address ||
                    printableForm.customer_address ||
                    "Same as billing address"}
                </small>
              </div>
            </div>
            <div className="invoice-print-context">
              <div>
                <span>Payment terms</span>
                <strong>{printableForm.terms || "—"}</strong>
              </div>
              <div>
                <span>Order / PO reference</span>
                <strong>{printableForm.order_number || "—"}</strong>
              </div>
              <div>
                <span>Salesperson</span>
                <strong>{printableForm.salesperson || "—"}</strong>
              </div>
              <div>
                <span>Subject</span>
                <strong>{printableForm.subject || "—"}</strong>
              </div>
            </div>
            <div className="invoice-print-items">
              <div className="invoice-print-items-heading">
                <span>Item details</span>
                <span>HSN / SAC</span>
                <span>Qty</span>
                <span>Rate</span>
                <span>GST</span>
                <span>Amount</span>
              </div>
              {printableForm.items.map((item, index) => (
                <div
                  className="invoice-print-items-row"
                  key={`${item.id || "item"}-${index}`}
                >
                  <span>
                    {item.item_details || "Unnamed item"}
                    <small>
                      {item.sku || "SKU —"}
                    </small>
                  </span>
                  <span>{item.hsn_code || "—"}</span>
                  <span>{item.quantity}</span>
                  <span>{money(item.rate)}</span>
                  <span>{item.gst_rate}%</span>
                  <span>
                    {money(numberValue(item.quantity) * numberValue(item.rate))}
                  </span>
                </div>
              ))}
            </div>
            <div className="invoice-print-lower">
              <div className="invoice-print-support">
                {printableForm.customer_notes && (
                  <div className="invoice-print-notes">
                    <span className="invoice-print-section-label">
                      Notes / terms &amp; conditions
                    </span>
                    <p>{printableForm.customer_notes}</p>
                  </div>
                )}
                {(bankName || bankAccount || bankIfsc || upiId) && (
                  <div className="invoice-print-payment">
                    <div className="invoice-print-payment-copy">
                      <span className="invoice-print-section-label">
                        Payment details
                      </span>
                      <strong>Scan to pay securely</strong>
                      {bankName && <small>Bank: {bankName}</small>}
                      {bankAccount && <small>Account: {bankAccount}</small>}
                      {bankIfsc && <small>IFSC: {bankIfsc}</small>}
                      {upiId && <small>UPI ID: {upiId}</small>}
                    </div>
                    {upiId && (
                      <img
                        src={`https://api.qrserver.com/v1/create-qr-code/?size=140x140&data=${encodeURIComponent(`upi://pay?pa=${upiId}&pn=${encodeURIComponent(sellerName)}&am=${printableQrAmount.toFixed(2)}&cu=INR`)}`}
                        alt={`UPI payment QR for ${upiId}`}
                      />
                    )}
                  </div>
                )}
              </div>
              <div className="invoice-print-summary">
                <span className="invoice-print-section-label">Invoice summary</span>
                {renderSummary(printableRecord)}
              </div>
            </div>
            <div className="invoice-print-total">
              <div>
                <span>Total amount</span>
                <small>Thank you for your business.</small>
              </div>
              <strong>{money(printableTotal)}</strong>
            </div>
            <footer className="invoice-print-footer">
              <span>Computer-generated invoice · No signature required</span>
              <span>{sellerName}</span>
            </footer>
          </article>
        )}
      </section>
    </div>
  );
}
