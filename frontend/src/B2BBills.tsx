import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { API_BASE } from './api';
import { useConfirm } from './ConfirmContext';
import { useToast } from './ToastContext';

const GST_STATE_MAP: Record<string, string> = {
	"01": "Jammu and Kashmir",
	"02": "Himachal Pradesh",
	"03": "Punjab",
	"04": "Chandigarh",
	"05": "Uttarakhand",
	"06": "Haryana",
	"07": "Delhi",
	"08": "Rajasthan",
	"09": "Uttar Pradesh",
	"10": "Bihar",
	"11": "Sikkim",
	"12": "Arunachal Pradesh",
	"13": "Nagaland",
	"14": "Manipur",
	"15": "Mizoram",
	"16": "Tripura",
	"17": "Meghalaya",
	"18": "Assam",
	"19": "West Bengal",
	"20": "Jharkhand",
	"21": "Odisha",
	"22": "Chhattisgarh",
	"23": "Madhya Pradesh",
	"24": "Gujarat",
	"26": "Dadra and Nagar Haveli and Daman and Diu",
	"27": "Maharashtra",
	"29": "Karnataka",
	"30": "Goa",
	"31": "Lakshadweep",
	"32": "Kerala",
	"33": "Tamil Nadu",
	"34": "Puducherry",
	"35": "Andaman and Nicobar Islands",
	"36": "Telangana",
	"37": "Andhra Pradesh",
	"38": "Ladakh"
};

function parseAddressString(addressStr: string) {
	if (!addressStr) return { street: '', city: '', pincode: '' };
	const parts = addressStr.split(',').map(s => s.trim());
	let pincode = '';
	let city = '';
	let streetParts = [...parts];

	const lastPart = parts[parts.length - 1] || '';
	const pinMatch = lastPart.match(/\b\d{6}\b/);
	if (pinMatch) {
		pincode = pinMatch[0];
		streetParts.pop();
	}

	if (streetParts.length >= 2) {
		const lastIndex = streetParts.length - 1;
		const possibleState = streetParts[lastIndex].toLowerCase();
		// If the second to last part looks like state candidate, pop it too
		const isState = Object.values(GST_STATE_MAP).some(s => s.toLowerCase() === possibleState);
		if (isState) {
			streetParts.pop();
		}
	}

	if (streetParts.length >= 1) {
		city = streetParts.pop() || '';
	}

	return {
		street: streetParts.join(', '),
		city: city,
		pincode: pincode
	};
}

interface B2BBillsProps {
	fetchWithAuth: (url: string, options?: RequestInit) => Promise<Response>;
	userRole?: string;
	appConfigs?: Record<string, string>;
}

interface B2BItem {
	id?: number;
	product_id?: number;
	item_details: string;
	sku?: string;
	hsn_code?: string;
	quantity: number;
	rate: number;
	amount: number;
	gst_rate?: number;
}

interface B2BInvoice {
	id?: number;
	invoice_number?: string;
	invoice_sequence?: number;
	financial_year?: string;
	order_number?: string;
	invoice_date: string;
	terms?: string;
	due_date?: string;
	salesperson?: string;
	subject?: string;
	customer_id?: number;
	customer_gstin: string;
	customer_name: string;
	customer_email?: string;
	customer_phone?: string;
	customer_state: string;
	customer_state_code: string;
	customer_address: string;
	customer_shipping_address?: string;
	seller_gstin?: string;
	seller_name?: string;
	seller_state?: string;
	seller_state_code?: string;
	seller_address?: string;
	subtotal_price: number;
	discount_percent: number;
	discount_amount: number;
	cgst_rate: number;
	cgst_amount: number;
	sgst_rate: number;
	sgst_amount: number;
	igst_rate: number;
	igst_amount: number;
	tds_tcs_type: string;
	tds_tcs_rate: number;
	tds_tcs_amount: number;
	transportation_charge: number;
	total_price: number;
	status: string;
	inventory_deducted?: boolean;
	payment_status: string;
	paid_amount: number;
	balance_amount: number;
	payment_date?: string;
	payment_method?: string;
	customer_notes?: string;
	items: B2BItem[];
}

interface B2BCustomer {
	id?: number;
	legal_name: string;
	trade_name?: string;
	gstin: string;
	pan?: string;
	email?: string;
	phone?: string;
	billing_address: string;
	shipping_address?: string;
	state: string;
	state_code: string;
	notes?: string;
}

interface B2BPaymentTerm {
	id?: number;
	name: string;
	due_days: number;
}

interface B2BInventoryItem {
	id: number;
	mi_sku: string;
	title: string;
	description: string;
	current_stock: number;
	price?: number;
	hsn_code?: string;
}

const DEFAULT_CUSTOMER_NOTES = `Thanks for your business.

Payment Terms: Full payment is required before the due date mentioned on the invoice. 

No Refunds & Returns: Due to the nature of our products, we do not accept returns or provide refunds once the item has been opened or used. If the product remains sealed and unused, you may contact us within 7 days for return eligibility, subject to approval.

Damaged or Incorrect Items: If you receive a damaged or incorrect product, please contact us within 48 hours of delivery with photographic evidence for a replacement or resolution.

Shipping & Delivery: We aim to deliver orders promptly, but delays due to courier services, customs, or unforeseen circumstances are beyond our control. Tracking details will be provided once your order is shipped.

Intellectual Property: All branding, packaging, and product names are trademarks of Millennial Perfumer™ and may not be reproduced without permission.`;

export function B2BBills({ fetchWithAuth, userRole = 'read', appConfigs = {} }: B2BBillsProps) {
	const { confirm: customConfirm } = useConfirm();
	const { success: toastSuccess, error: toastError } = useToast();

	// Navigation State
	const [activeSubTab, setActiveSubTab] = useState<'invoices' | 'proformas' | 'credit-notes' | 'debit-notes' | 'customers' | 'outstanding' | 'locks'>('invoices');
	const [viewMode, setViewMode] = useState<'list' | 'create' | 'edit' | 'preview' | 'create-cn' | 'edit-cn' | 'preview-cn' | 'create-dn' | 'edit-dn' | 'preview-dn' | 'list-pf' | 'create-pf' | 'edit-pf' | 'preview-pf'>('list');

	// Inventory products state
	const [inventoryProducts, setInventoryProducts] = useState<B2BInventoryItem[]>([]);

	// Invoices List state
	const [invoices, setInvoices] = useState<B2BInvoice[]>([]);
	const [invoiceSearch, setInvoiceSearch] = useState('');
	const [invoiceStatusFilter, setInvoiceStatusFilter] = useState('');
	const [selectedInvoice, setSelectedInvoice] = useState<B2BInvoice | null>(null);

	// Proformas List state
	const [proformas, setProformas] = useState<any[]>([]);
	const [proformaSearch, setProformaSearch] = useState('');
	const [proformaStatusFilter, setProformaStatusFilter] = useState('');
	const [selectedProforma, setSelectedProforma] = useState<any | null>(null);
	const [, setNextProformaNumber] = useState<string>('');

	// Creator / Editor Form State for Proforma
	const [formProforma, setFormProforma] = useState<any>({
		note_date: new Date().toISOString().split('T')[0],
		valid_until: '',
		customer_gstin: '',
		customer_name: '',
		customer_state: '',
		customer_state_code: '',
		customer_address: '',
		customer_shipping_address: '',
		subtotal_price: 0,
		discount_percent: 0,
		discount_amount: 0,
		cgst_rate: 0,
		cgst_amount: 0,
		sgst_rate: 0,
		sgst_amount: 0,
		igst_rate: 0,
		igst_amount: 0,
		total_price: 0,
		advance_paid: 0,
		status: 'DRAFT',
		revision_number: 1,
		items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
	});

	// Credit & Debit Notes lists & search
	const [creditNotes, setCreditNotes] = useState<any[]>([]);
	const [debitNotes, setDebitNotes] = useState<any[]>([]);
	const [creditSearch, setCreditSearch] = useState('');
	const [debitSearch, setDebitSearch] = useState('');
	const [selectedCreditNote, setSelectedCreditNote] = useState<any | null>(null);
	const [selectedDebitNote, setSelectedDebitNote] = useState<any | null>(null);

	// Customers List state
	const [customers, setCustomers] = useState<B2BCustomer[]>([]);
	const [customerSearch, setCustomerSearch] = useState('');
	const [showCustomerModal, setShowCustomerModal] = useState(false);
	const [editingCustomer, setEditingCustomer] = useState<B2BCustomer | null>(null);
	const [sameAsBilling, setSameAsBilling] = useState(true);

	// Address breakdown states
	const [billingStreet, setBillingStreet] = useState('');
	const [billingCity, setBillingCity] = useState('');
	const [billingPincode, setBillingPincode] = useState('');

	const [shippingStreet, setShippingStreet] = useState('');
	const [shippingCity, setShippingCity] = useState('');
	const [shippingPincode, setShippingPincode] = useState('');

	// Payment Terms state
	const [paymentTerms, setPaymentTerms] = useState<B2BPaymentTerm[]>([]);
	const [showPaymentTermModal, setShowPaymentTermModal] = useState(false);
	const [newTermName, setNewTermName] = useState('');
	const [newTermDays, setNewTermDays] = useState(0);

	// Customer Ledger details modal state
	const [ledgerCustomer, setLedgerCustomer] = useState<B2BCustomer | null>(null);
	const [ledgerData, setLedgerData] = useState<any | null>(null);
	const [showLedgerModal, setShowLedgerModal] = useState(false);

	// Outstanding Aging Report state
	const [outstandingReport, setOutstandingReport] = useState<any[]>([]);

	// Custom Confirm modal state
	const [confirmConfig, setConfirmConfig] = useState<{
		show: boolean;
		message: string;
		resolve?: (val: boolean) => void;
	}>({ show: false, message: '' });

	const showConfirm = (message: string): Promise<boolean> => {
		return new Promise((resolve) => {
			setConfirmConfig({
				show: true,
				message,
				resolve
			});
		});
	};

	// Custom Alert modal state
	const [alertConfig, setAlertConfig] = useState<{
		show: boolean;
		message: string;
	}>({ show: false, message: '' });

	const alert = (message: string) => {
		setAlertConfig({ show: true, message });
	};

	const getProformaStatusStyle = (status: string) => {
		switch (status) {
			case 'DRAFT':
				return { bg: 'rgba(156, 163, 175, 0.1)', color: '#4b5563', border: 'rgba(156, 163, 175, 0.2)' };
			case 'SENT':
				return { bg: 'rgba(245, 158, 11, 0.1)', color: '#d97706', border: 'rgba(245, 158, 11, 0.2)' };
			case 'ACCEPTED':
				return { bg: 'rgba(16, 185, 129, 0.1)', color: '#059669', border: 'rgba(16, 185, 129, 0.2)' };
			case 'CONVERTED_TO_INVOICE':
				return { bg: 'rgba(99, 102, 241, 0.1)', color: '#4f46e5', border: 'rgba(99, 102, 241, 0.2)' };
			case 'REJECTED':
				return { bg: 'rgba(239, 68, 68, 0.1)', color: '#dc2626', border: 'rgba(239, 68, 68, 0.2)' };
			case 'CANCELLED':
				return { bg: 'rgba(220, 38, 38, 0.15)', color: '#b91c1c', border: 'rgba(220, 38, 38, 0.2)' };
			case 'EXPIRED':
				return { bg: 'rgba(107, 114, 128, 0.1)', color: '#374151', border: 'rgba(107, 114, 128, 0.2)' };
			default:
				return { bg: 'rgba(156, 163, 175, 0.1)', color: '#4b5563', border: 'rgba(156, 163, 175, 0.2)' };
		}
	};

	// GST locks/periods state
	const [gstPeriods, setGstPeriods] = useState<any[]>([]);

	// Credit & Debit Note Form States
	const [formCreditNote, setFormCreditNote] = useState<any>({
		note_date: new Date().toISOString().split('T')[0],
		customer_gstin: '',
		customer_name: '',
		customer_state: '',
		customer_state_code: '',
		customer_address: '',
		subtotal_price: 0,
		discount_percent: 0,
		discount_amount: 0,
		cgst_rate: 0,
		cgst_amount: 0,
		sgst_rate: 0,
		sgst_amount: 0,
		igst_rate: 0,
		igst_amount: 0,
		total_price: 0,
		status: 'DRAFT',
		reason: '',
		invoice_id: undefined,
		items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
	});

	const [formDebitNote, setFormDebitNote] = useState<any>({
		note_date: new Date().toISOString().split('T')[0],
		customer_gstin: '',
		customer_name: '',
		customer_state: '',
		customer_state_code: '',
		customer_address: '',
		subtotal_price: 0,
		discount_percent: 0,
		discount_amount: 0,
		cgst_rate: 0,
		cgst_amount: 0,
		sgst_rate: 0,
		sgst_amount: 0,
		igst_rate: 0,
		igst_amount: 0,
		total_price: 0,
		status: 'DRAFT',
		reason: '',
		invoice_id: undefined,
		items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
	});

	const validateGSTINFormat = (gstin: string) => {
		const regex = /^\d{2}[A-Z]{5}\d{4}[A-Z]{1}[A-Z\d]{1}[Z]{1}[A-Z\d]{1}$/;
		return regex.test(gstin);
	};

	useEffect(() => {
		if (editingCustomer) {
			const parsedBilling = parseAddressString(editingCustomer.billing_address);
			setBillingStreet(parsedBilling.street);
			setBillingCity(parsedBilling.city);
			setBillingPincode(parsedBilling.pincode);

			const parsedShipping = parseAddressString(editingCustomer.shipping_address || '');
			setShippingStreet(parsedShipping.street);
			setShippingCity(parsedShipping.city);
			setShippingPincode(parsedShipping.pincode);
		} else {
			setBillingStreet('');
			setBillingCity('');
			setBillingPincode('');
			setShippingStreet('');
			setShippingCity('');
			setShippingPincode('');
		}
	}, [editingCustomer]);

	// Payment Log modal
	const [showPaymentModal, setShowPaymentModal] = useState(false);
	const [paymentInvoice, setPaymentInvoice] = useState<B2BInvoice | null>(null);
	const [paymentAmount, setPaymentAmount] = useState(0);
	const [paymentMethod, setPaymentMethod] = useState('Bank Transfer');

	// Next invoice number preview (shown in create/edit form header)
	const [nextInvoiceNumber, setNextInvoiceNumber] = useState<string>('');

	// Creator / Editor Form State
	const [formInvoice, setFormInvoice] = useState<B2BInvoice>({
		invoice_date: new Date().toISOString().split('T')[0],
		customer_gstin: '',
		customer_name: '',
		customer_state: '',
		customer_state_code: '',
		customer_address: '',
		customer_shipping_address: '',
		subtotal_price: 0,
		discount_percent: 0,
		discount_amount: 0,
		cgst_rate: 0,
		cgst_amount: 0,
		sgst_rate: 0,
		sgst_amount: 0,
		igst_rate: 0,
		igst_amount: 0,
		tds_tcs_type: 'NONE',
		tds_tcs_rate: 0,
		tds_tcs_amount: 0,
		transportation_charge: 0,
		total_price: 0,
		status: 'DRAFT',
		payment_status: 'UNPAID',
		paid_amount: 0,
		balance_amount: 0,
		customer_notes: DEFAULT_CUSTOMER_NOTES,
		items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
	});

	const [isEditingBilling, setIsEditingBilling] = useState(false);
	const [isEditingShipping, setIsEditingShipping] = useState(false);

	// Load Data
	useEffect(() => {
		loadInvoices();
		loadProformas();
		loadCustomers();
		loadInventoryProducts();
		loadPaymentTerms();
		loadCreditNotes();
		loadDebitNotes();
		loadOutstandingReport();
		loadGSTPeriods();
	}, []);

	const loadInvoices = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices`);
			if (res.ok) {
				const data = await res.json();
				setInvoices(data || []);
			}
		} catch (err) {
			console.error('Failed to load invoices:', err);
		}
	};

	const loadProformas = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas`);
			if (res.ok) {
				const data = await res.json();
				setProformas(data || []);
			}
		} catch (err) {
			console.error('Failed to load proformas:', err);
		}
	};

	const fetchNextProformaNumber = async (date?: string) => {
		try {
			const d = date || new Date().toISOString().split('T')[0];
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas/next-number?date=${d}`);
			if (res.ok) {
				const data = await res.json();
				setNextProformaNumber(data.next_proforma_number || '');
			}
		} catch (err) {
			console.error('Failed to fetch next proforma number:', err);
		}
	};

	const handleSaveProforma = async (asDraft: boolean) => {
		if (formProforma.customer_gstin && !validateGSTINFormat(formProforma.customer_gstin)) {
			alert('Warning: Customer GSTIN format is invalid. Standard format: 22AAAAA1111A1Z1');
			return;
		}
		try {
			const pf = { ...formProforma } as any;
			if (pf.note_date) {
				pf.note_date = new Date(pf.note_date).toISOString();
			}
			if (pf.valid_until) {
				pf.valid_until = new Date(pf.valid_until).toISOString();
			} else {
				delete pf.valid_until;
			}
			const method = (viewMode === 'edit-pf' || formProforma.id || pf.id) ? 'PUT' : 'POST';
			if (method === 'PUT' && !pf.id && formProforma.id) {
				pf.id = formProforma.id;
			}
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas`, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(pf)
			});
			if (res.ok) {
				const savedPf = await res.json();
				if (!asDraft) {
					const issueRes = await fetchWithAuth(`${API_BASE}/api/b2b/proformas/issue?id=${savedPf.id}`, {
						method: 'POST'
					});
					if (!issueRes.ok) {
						const text = await issueRes.text();
						alert('Saved as draft, but activation failed: ' + text);
					}
				}
				setViewMode('list');
				loadProformas();
			} else {
				const text = await res.text();
				alert(text || 'Failed to save proforma');
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving proforma');
		}
	};

	const handleDeleteProforma = async (id: number) => {
		if (!await showConfirm('Are you sure you want to delete this proforma? This action cannot be undone.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas?id=${id}`, {
				method: 'DELETE'
			});
			if (res.ok) {
				loadProformas();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleIssueProforma = async (id: number) => {
		if (!await showConfirm('Are you sure you want to issue this proforma invoice? This locks modifications and generates the sequential number.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas/issue?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadProformas();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleAcceptProforma = async (id: number) => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas/accept?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadProformas();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};


	const handleCancelProforma = async (id: number) => {
		if (!await showConfirm('Are you sure you want to cancel this proforma?')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas/cancel?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadProformas();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleCreateRevision = async (id: number) => {
		if (!await showConfirm('Are you sure you want to create a new revision of this proforma?')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas/revision?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				const newPf = await res.json();
				alert('Revision draft created successfully.');
				loadProformas();
				setFormProforma(newPf);
				setViewMode('edit-pf');
				fetchNextProformaNumber(newPf.note_date);
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleConvertToTaxInvoice = async (id: number) => {
		if (!await showConfirm('Are you sure you want to convert this proforma into a Tax Invoice? This will create a draft tax invoice and mark the proforma as converted.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas/convert?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				alert('Successfully converted to draft Tax Invoice.');
				loadProformas();
				loadInvoices();
				setActiveSubTab('invoices');
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleCheckExpiredProformas = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/proformas/check-expiry`, {
				method: 'POST'
			});
			if (res.ok) {
				const data = await res.json();
				if (data.expired_count > 0) {
					alert(`Checked and marked ${data.expired_count} proforma(s) as expired.`);
					loadProformas();
				} else {
					alert('No proformas found that have expired.');
				}
			}
		} catch (err) {
			console.error(err);
		}
	};

	const getProductStock = (productId?: number) => {
		if (!productId) return null;
		const prod = inventoryProducts.find(p => p.id === productId);
		return prod ? prod.current_stock : null;
	};

	const loadInventoryProducts = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/inventory`);
			if (res.ok) {
				const data = await res.json();
				setInventoryProducts(data || []);
			}
		} catch (err) {
			console.error('Failed to load inventory products:', err);
		}
	};

	const loadCustomers = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers`);
			if (res.ok) {
				const data = await res.json();
				setCustomers(data || []);
			}
		} catch (err) {
			console.error('Failed to load customers:', err);
		}
	};

	const loadCreditNotes = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/credit-notes`);
			if (res.ok) {
				const data = await res.json();
				setCreditNotes(data || []);
			}
		} catch (err) {
			console.error('Failed to load credit notes:', err);
		}
	};

	const loadDebitNotes = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/debit-notes`);
			if (res.ok) {
				const data = await res.json();
				setDebitNotes(data || []);
			}
		} catch (err) {
			console.error('Failed to load debit notes:', err);
		}
	};

	const loadOutstandingReport = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers/outstanding`);
			if (res.ok) {
				const data = await res.json();
				setOutstandingReport(data || []);
			}
		} catch (err) {
			console.error('Failed to load outstanding report:', err);
		}
	};

	const loadGSTPeriods = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/gst-periods`);
			if (res.ok) {
				const data = await res.json();
				setGstPeriods(data || []);
			}
		} catch (err) {
			console.error('Failed to load GST periods:', err);
		}
	};

	const loadCustomerLedger = async (customerId: number) => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers/ledger?customer_id=${customerId}`);
			if (res.ok) {
				const data = await res.json();
				setLedgerData(data);
			}
		} catch (err) {
			console.error('Failed to load customer ledger:', err);
		}
	};

	const toggleGSTPeriod = async (month: number, year: number, currentStatus: string) => {
		try {
			const newStatus = currentStatus === 'LOCKED' ? 'OPEN' : 'LOCKED';
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/gst-periods`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ month, year, status: newStatus })
			});
			if (res.ok) {
				loadGSTPeriods();
			} else {
				const text = await res.text();
				alert(text || 'Failed to toggle lock status');
			}
		} catch (err) {
			console.error(err);
		}
	};

	const fetchNextInvoiceNumber = async (date?: string) => {
		try {
			const d = date || new Date().toISOString().split('T')[0];
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/next-number?date=${d}`);
			if (res.ok) {
				const data = await res.json();
				setNextInvoiceNumber(data.next_invoice_number || '');
			}
		} catch (err) {
			console.error('Failed to fetch next invoice number:', err);
		}
	};

	const loadPaymentTerms = async () => {
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/payment-terms`);
			if (res.ok) {
				const data = await res.json();
				setPaymentTerms(data || []);
			}
		} catch (err) {
			console.error('Failed to load payment terms:', err);
		}
	};

	const handleTermsChange = (termName: string, currentInvoiceDate?: string) => {
		const invoiceDate = currentInvoiceDate || formInvoice.invoice_date || new Date().toISOString().split('T')[0];
		const matchedTerm = paymentTerms.find(t => t.name === termName);

		let computedDueDate = '';
		if (matchedTerm) {
			const date = new Date(invoiceDate);
			date.setDate(date.getDate() + matchedTerm.due_days);
			computedDueDate = date.toISOString().split('T')[0];
		}

		setFormInvoice({
			...formInvoice,
			terms: termName,
			due_date: computedDueDate
		});
	};

	const handleSavePaymentTerm = async () => {
		if (!newTermName.trim()) {
			alert('Term name is required');
			return;
		}
		if (newTermDays < 0) {
			alert('Due Days cannot be negative');
			return;
		}

		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/payment-terms`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name: newTermName.trim(), due_days: Number(newTermDays) })
			});
			if (res.ok) {
				const term = await res.json();
				setNewTermName('');
				setNewTermDays(0);
				setShowPaymentTermModal(false);

				// Reload the terms list
				const refreshRes = await fetchWithAuth(`${API_BASE}/api/b2b/payment-terms`);
				if (refreshRes.ok) {
					const data = await refreshRes.json();
					setPaymentTerms(data || []);

					// Automatically select the newly created term and compute its due date
					const matchedTerm = (data || []).find((t: any) => t.name === term.name);
					const invoiceDate = formInvoice.invoice_date || new Date().toISOString().split('T')[0];
					let computedDueDate = '';
					if (matchedTerm) {
						const date = new Date(invoiceDate);
						date.setDate(date.getDate() + matchedTerm.due_days);
						computedDueDate = date.toISOString().split('T')[0];
					}
					setFormInvoice({
						...formInvoice,
						terms: term.name,
						due_date: computedDueDate
					});
				}
			} else {
				const text = await res.text();
				alert('Failed to save payment term: ' + text);
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving payment term');
		}
	};

	// Save customer
	const handleSaveCustomer = async (cust: B2BCustomer) => {
		if (cust.gstin && !validateGSTINFormat(cust.gstin)) {
			alert('Warning: Customer GSTIN format is invalid. Standard format: 22AAAAA1111A1Z1');
			return;
		}
		try {
			const finalCust = { ...cust };

			// Assemble Billing Address
			const billingParts = [];
			if (billingStreet.trim()) billingParts.push(billingStreet.trim());
			if (billingCity.trim()) billingParts.push(billingCity.trim());
			if (cust.state.trim()) billingParts.push(cust.state.trim());
			if (billingPincode.trim()) billingParts.push(billingPincode.trim());

			finalCust.billing_address = billingParts.join(', ');

			if (sameAsBilling) {
				finalCust.shipping_address = finalCust.billing_address;
			} else {
				// Assemble Shipping Address
				const shippingParts = [];
				if (shippingStreet.trim()) shippingParts.push(shippingStreet.trim());
				if (shippingCity.trim()) shippingParts.push(shippingCity.trim());
				if (cust.state.trim()) shippingParts.push(cust.state.trim());
				if (shippingPincode.trim()) shippingParts.push(shippingPincode.trim());

				finalCust.shipping_address = shippingParts.join(', ');
			}
			const method = finalCust.id ? 'PUT' : 'POST';
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers`, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(finalCust)
			});
			if (res.ok) {
				setShowCustomerModal(false);
				setEditingCustomer(null);
				loadCustomers();
			} else {
				const text = await res.text();
				alert(text || 'Failed to save customer');
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving customer');
		}
	};

	// Delete Customer
	const handleDeleteCustomer = async (id: number) => {
		if (!await showConfirm('Are you sure you want to delete this customer?')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/customers?id=${id}`, {
				method: 'DELETE'
			});
			if (res.ok) {
				loadCustomers();
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Save B2B Invoice (Draft / Active)
	const handleSaveInvoice = async (asDraft: boolean) => {
		if (formInvoice.customer_gstin && !validateGSTINFormat(formInvoice.customer_gstin)) {
			alert('Warning: Customer GSTIN format is invalid. Standard format: 22AAAAA1111A1Z1');
			return;
		}
		try {
			const inv = { ...formInvoice } as any;
			if (inv.invoice_date) {
				inv.invoice_date = new Date(inv.invoice_date).toISOString();
			}
			if (inv.due_date) {
				inv.due_date = new Date(inv.due_date).toISOString();
			} else {
				delete inv.due_date;
			}
			if (inv.payment_date) {
				inv.payment_date = new Date(inv.payment_date).toISOString();
			} else {
				delete inv.payment_date;
			}
			
			// Remove other potentially empty fields
			if (inv.order_number === '') delete inv.order_number;
			if (inv.terms === '') delete inv.terms;
			if (inv.salesperson === '') delete inv.salesperson;
			if (inv.subject === '') delete inv.subject;

			const method = inv.id ? 'PUT' : 'POST';

			// Save Invoice first
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices`, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(inv)
			});

			if (res.ok) {
				const savedInv = await res.json();

				// If clicking "Save and Send" (Activate), trigger Issue API
				if (!asDraft) {
					const issueRes = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/issue?id=${savedInv.id}`, {
						method: 'POST'
					});
					if (!issueRes.ok) {
						const text = await issueRes.text();
						alert('Saved as draft, but activation failed: ' + text);
					}
				}

				setViewMode('list');
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text || 'Failed to save invoice');
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving invoice');
		}
	};

	// Delete Invoice
	const handleDeleteInvoice = async (id: number) => {
		const inv = invoices.find(i => i.id === id);
		const isDraft = inv?.status === 'DRAFT';
		const confirmMsg = isDraft
			? 'Are you sure you want to delete this draft invoice?'
			: `WARNING: Deleting an active or cancelled invoice (${inv?.invoice_number || 'ID: ' + id}) can impact tax reports and calculations. Are you sure you want to permanently delete this invoice?`;

		const confirmDelete = await customConfirm({
			title: isDraft ? 'Delete Draft Invoice' : 'Delete Active/Cancelled Invoice',
			message: confirmMsg,
			confirmLabel: 'Delete',
			cancelLabel: 'Cancel',
			variant: 'danger'
		});

		if (!confirmDelete) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices?id=${id}`, {
				method: 'DELETE'
			});
			if (res.ok) {
				toastSuccess('Invoice deleted successfully');
				loadInvoices();
			} else {
				const text = await res.text();
				toastError(text || 'Failed to delete invoice');
			}
		} catch (err) {
			console.error(err);
			toastError('Network error deleting invoice');
		}
	};

	// Issue Draft Invoice
	const handleIssueInvoice = async (id: number) => {
		const confirmIssue = await customConfirm({
			title: 'Issue / Activate Invoice',
			message: 'Are you sure you want to activate/issue this invoice? This locks modifications and generates the sequential B2B invoice number.',
			confirmLabel: 'Issue Invoice',
			cancelLabel: 'Cancel',
			variant: 'primary'
		});

		if (!confirmIssue) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/issue?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				toastSuccess('Invoice issued successfully');
				loadInvoices();
			} else {
				const text = await res.text();
				toastError(text || 'Failed to issue invoice');
			}
		} catch (err) {
			console.error(err);
			toastError('Network error issuing invoice');
		}
	};

	// Cancel issued invoice
	const handleCancelInvoice = async (id: number) => {
		const confirmCancel = await customConfirm({
			title: 'Cancel Issued Invoice',
			message: 'Are you sure you want to CANCEL this issued invoice? This removes it from active revenue and tax summaries historically.',
			confirmLabel: 'Cancel Invoice',
			cancelLabel: 'Keep Active',
			variant: 'danger'
		});

		if (!confirmCancel) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/cancel?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				toastSuccess('Invoice cancelled successfully');
				loadInvoices();
			} else {
				const text = await res.text();
				toastError(text || 'Failed to cancel invoice');
			}
		} catch (err) {
			console.error(err);
			toastError('Network error cancelling invoice');
		}
	};

	// Deduct inventory for issued invoice
	const handleDeductInventory = async (id: number) => {
		const confirmDeduct = await customConfirm({
			title: 'Deduct Physical Stock',
			message: 'Are you sure you want to deduct physical stock for this invoice? This will decrease warehouse quantities for all items in this invoice.',
			confirmLabel: 'Deduct Stock',
			cancelLabel: 'Cancel',
			variant: 'primary'
		});

		if (!confirmDeduct) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/deduct-inventory?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				toastSuccess('Inventory deducted successfully!');
				loadInvoices();
				// Also update the selected invoice preview state if open
				setSelectedInvoice(prev => {
					if (prev && prev.id === id) {
						return { ...prev, inventory_deducted: true };
					}
					return prev;
				});
			} else {
				const text = await res.text();
				toastError(text || 'Failed to deduct inventory');
			}
		} catch (err) {
			console.error(err);
			toastError('Failed to deduct inventory');
		}
	};

	// Revert inventory deduction for issued invoice
	const handleRevertInventory = async (id: number) => {
		const confirmRevert = await customConfirm({
			title: 'Revert Stock Deduction',
			message: 'Are you sure you want to REVERT the stock deduction for this invoice? This will add back the quantities to the warehouse for all items.',
			confirmLabel: 'Revert Stock',
			cancelLabel: 'Cancel',
			variant: 'danger'
		});

		if (!confirmRevert) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/revert-inventory?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				toastSuccess('Stock reverted successfully!');
				loadInvoices();
				// Also update the selected invoice preview state if open
				setSelectedInvoice(prev => {
					if (prev && prev.id === id) {
						return { ...prev, inventory_deducted: false };
					}
					return prev;
				});
			} else {
				const text = await res.text();
				toastError(text || 'Failed to revert stock');
			}
		} catch (err) {
			console.error(err);
			toastError('Failed to revert stock');
		}
	};

	// Save Payment details
	const handleSavePayment = async () => {
		if (!paymentInvoice) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/invoices/payment`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					id: paymentInvoice.id,
					paid_amount: paymentAmount,
					payment_method: paymentMethod
				})
			});
			if (res.ok) {
				setShowPaymentModal(false);
				setPaymentInvoice(null);
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Credit Note CRUD Actions
	const handleSaveCreditNote = async (asDraft: boolean) => {
		if (formCreditNote.customer_gstin && !validateGSTINFormat(formCreditNote.customer_gstin)) {
			alert('Warning: Customer GSTIN format is invalid. Standard format: 22AAAAA1111A1Z1');
			return;
		}
		try {
			const cn = { ...formCreditNote };
			const method = cn.id ? 'PUT' : 'POST';
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/credit-notes`, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					...cn,
					note_date: new Date(cn.note_date).toISOString()
				})
			});
			if (res.ok) {
				const savedNote = await res.json();
				if (!asDraft) {
					const issueRes = await fetchWithAuth(`${API_BASE}/api/b2b/credit-notes/issue?id=${savedNote.id}`, {
						method: 'POST'
					});
					if (!issueRes.ok) {
						const text = await issueRes.text();
						alert('Saved as draft, but activation failed: ' + text);
					}
				}
				setViewMode('list');
				loadCreditNotes();
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text || 'Failed to save credit note');
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving credit note');
		}
	};

	const handleDeleteCreditNote = async (id: number) => {
		if (!await showConfirm('Are you sure you want to delete this credit note?')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/credit-notes?id=${id}`, {
				method: 'DELETE'
			});
			if (res.ok) {
				loadCreditNotes();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleIssueCreditNote = async (id: number) => {
		if (!await showConfirm('Are you sure you want to issue this credit note? This will post it and adjust the linked invoice outstanding balance.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/credit-notes/issue?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadCreditNotes();
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleCancelCreditNote = async (id: number) => {
		if (!await showConfirm('Are you sure you want to cancel this credit note? This will restore the outstanding balance of the linked invoice.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/credit-notes/cancel?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadCreditNotes();
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Debit Note CRUD Actions
	const handleSaveDebitNote = async (asDraft: boolean) => {
		if (formDebitNote.customer_gstin && !validateGSTINFormat(formDebitNote.customer_gstin)) {
			alert('Warning: Customer GSTIN format is invalid. Standard format: 22AAAAA1111A1Z1');
			return;
		}
		try {
			const dn = { ...formDebitNote };
			const method = dn.id ? 'PUT' : 'POST';
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/debit-notes`, {
				method,
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					...dn,
					note_date: new Date(dn.note_date).toISOString()
				})
			});
			if (res.ok) {
				const savedNote = await res.json();
				if (!asDraft) {
					const issueRes = await fetchWithAuth(`${API_BASE}/api/b2b/debit-notes/issue?id=${savedNote.id}`, {
						method: 'POST'
					});
					if (!issueRes.ok) {
						const text = await issueRes.text();
						alert('Saved as draft, but activation failed: ' + text);
					}
				}
				setViewMode('list');
				loadDebitNotes();
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text || 'Failed to save debit note');
			}
		} catch (err) {
			console.error(err);
			alert('Network error saving debit note');
		}
	};

	const handleDeleteDebitNote = async (id: number) => {
		if (!await showConfirm('Are you sure you want to delete this debit note?')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/debit-notes?id=${id}`, {
				method: 'DELETE'
			});
			if (res.ok) {
				loadDebitNotes();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleIssueDebitNote = async (id: number) => {
		if (!await showConfirm('Are you sure you want to issue this debit note? This will post it and adjust the linked invoice outstanding balance.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/debit-notes/issue?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadDebitNotes();
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	const handleCancelDebitNote = async (id: number) => {
		if (!await showConfirm('Are you sure you want to cancel this debit note? This will reduce the outstanding balance of the linked invoice.')) return;
		try {
			const res = await fetchWithAuth(`${API_BASE}/api/b2b/debit-notes/cancel?id=${id}`, {
				method: 'POST'
			});
			if (res.ok) {
				loadDebitNotes();
				loadInvoices();
			} else {
				const text = await res.text();
				alert(text);
			}
		} catch (err) {
			console.error(err);
		}
	};

	// Dynamic Totals calculation helper in form
	const recalculateTotals = (updatedInvoice: B2BInvoice) => {
		let subtotal = 0;
		updatedInvoice.items.forEach(item => {
			item.amount = (item.quantity || 0) * (item.rate || 0);
			subtotal += item.amount;
		});
		updatedInvoice.subtotal_price = subtotal;

		if (updatedInvoice.discount_percent > 0) {
			updatedInvoice.discount_amount = (subtotal * updatedInvoice.discount_percent) / 100;
		} else {
			updatedInvoice.discount_amount = 0;
		}

		const taxable = subtotal - updatedInvoice.discount_amount;
		const discountRatio = subtotal > 0 ? taxable / subtotal : 1;

		// Reset tax values
		updatedInvoice.cgst_rate = 0;
		updatedInvoice.cgst_amount = 0;
		updatedInvoice.sgst_rate = 0;
		updatedInvoice.sgst_amount = 0;
		updatedInvoice.igst_rate = 0;
		updatedInvoice.igst_amount = 0;

		const isSameState = updatedInvoice.customer_state_code === '33'; // TN Seller default matching prefix '33'

		updatedInvoice.items.forEach(item => {
			const itemSubtotal = (item.quantity || 0) * (item.rate || 0);
			const itemTaxable = itemSubtotal * discountRatio;
			const itemGstRate = item.gst_rate !== undefined ? item.gst_rate : 18;

			if (isSameState) {
				const cgstRate = itemGstRate / 2;
				const sgstRate = itemGstRate / 2;
				updatedInvoice.cgst_amount += (itemTaxable * cgstRate) / 100;
				updatedInvoice.sgst_amount += (itemTaxable * sgstRate) / 100;
				// Maintain active rates for summary presentation
				updatedInvoice.cgst_rate = cgstRate;
				updatedInvoice.sgst_rate = sgstRate;
			} else {
				const igstRate = itemGstRate;
				updatedInvoice.igst_amount += (itemTaxable * igstRate) / 100;
				updatedInvoice.igst_rate = igstRate;
			}
		});

		const totalTax = updatedInvoice.cgst_amount + updatedInvoice.sgst_amount + updatedInvoice.igst_amount;

		updatedInvoice.tds_tcs_amount = 0;
		if (updatedInvoice.tds_tcs_type !== 'NONE') {
			updatedInvoice.tds_tcs_amount = (taxable * (updatedInvoice.tds_tcs_rate || 0)) / 100;
		}

		let finalTotal = taxable + totalTax + (Number(updatedInvoice.transportation_charge) || 0);
		if (updatedInvoice.tds_tcs_type === 'TCS') {
			finalTotal += updatedInvoice.tds_tcs_amount;
		} else if (updatedInvoice.tds_tcs_type === 'TDS') {
			finalTotal -= updatedInvoice.tds_tcs_amount;
		}

		updatedInvoice.total_price = finalTotal;
		updatedInvoice.balance_amount = finalTotal - (updatedInvoice.paid_amount || 0);

		setFormInvoice({ ...updatedInvoice });
	};

	const recalculateProformaTotals = (updatedPf: any) => {
		let subtotal = 0;
		updatedPf.items.forEach((item: any) => {
			item.amount = (item.quantity || 0) * (item.rate || 0);
			subtotal += item.amount;
		});
		updatedPf.subtotal_price = subtotal;

		if (updatedPf.discount_percent > 0) {
			updatedPf.discount_amount = (subtotal * updatedPf.discount_percent) / 100;
		} else {
			updatedPf.discount_amount = 0;
		}

		const taxable = subtotal - updatedPf.discount_amount;
		const discountRatio = subtotal > 0 ? taxable / subtotal : 1;

		updatedPf.cgst_rate = 0;
		updatedPf.cgst_amount = 0;
		updatedPf.sgst_rate = 0;
		updatedPf.sgst_amount = 0;
		updatedPf.igst_rate = 0;
		updatedPf.igst_amount = 0;

		const isSameState = updatedPf.customer_state_code === '33';

		updatedPf.items.forEach((item: any) => {
			const itemSubtotal = (item.quantity || 0) * (item.rate || 0);
			const itemTaxable = itemSubtotal * discountRatio;
			const itemGstRate = item.gst_rate !== undefined ? item.gst_rate : 18;

			if (isSameState) {
				const cgstRate = itemGstRate / 2;
				const sgstRate = itemGstRate / 2;
				updatedPf.cgst_amount += (itemTaxable * cgstRate) / 100;
				updatedPf.sgst_amount += (itemTaxable * sgstRate) / 100;
				updatedPf.cgst_rate = cgstRate;
				updatedPf.sgst_rate = sgstRate;
			} else {
				const igstRate = itemGstRate;
				updatedPf.igst_amount += (itemTaxable * igstRate) / 100;
				updatedPf.igst_rate = igstRate;
			}
		});

		const totalTax = updatedPf.cgst_amount + updatedPf.sgst_amount + updatedPf.igst_amount;
		updatedPf.total_price = taxable + totalTax;

		setFormProforma({ ...updatedPf });
	};

	const recalculateCreditNoteTotals = (updatedNote: any) => {
		let subtotal = 0;
		updatedNote.items.forEach((item: any) => {
			item.amount = (item.quantity || 0) * (item.rate || 0);
			subtotal += item.amount;
		});
		updatedNote.subtotal_price = subtotal;

		if (updatedNote.discount_percent > 0) {
			updatedNote.discount_amount = (subtotal * updatedNote.discount_percent) / 100;
		} else {
			updatedNote.discount_amount = 0;
		}

		const taxable = subtotal - updatedNote.discount_amount;

		// Reset tax values
		updatedNote.cgst_rate = 0;
		updatedNote.cgst_amount = 0;
		updatedNote.sgst_rate = 0;
		updatedNote.sgst_amount = 0;
		updatedNote.igst_rate = 0;
		updatedNote.igst_amount = 0;

		const isSameState = updatedNote.customer_state_code === '33';
		const defaultTaxRate = 18;

		if (isSameState) {
			updatedNote.cgst_rate = defaultTaxRate / 2;
			updatedNote.sgst_rate = defaultTaxRate / 2;
			updatedNote.cgst_amount = (taxable * updatedNote.cgst_rate) / 100;
			updatedNote.sgst_amount = (taxable * updatedNote.sgst_rate) / 100;
		} else {
			updatedNote.igst_rate = defaultTaxRate;
			updatedNote.igst_amount = (taxable * updatedNote.igst_rate) / 100;
		}

		updatedNote.total_price = taxable + updatedNote.cgst_amount + updatedNote.sgst_amount + updatedNote.igst_amount;
		setFormCreditNote({ ...updatedNote });
	};

	const recalculateDebitNoteTotals = (updatedNote: any) => {
		let subtotal = 0;
		updatedNote.items.forEach((item: any) => {
			item.amount = (item.quantity || 0) * (item.rate || 0);
			subtotal += item.amount;
		});
		updatedNote.subtotal_price = subtotal;

		if (updatedNote.discount_percent > 0) {
			updatedNote.discount_amount = (subtotal * updatedNote.discount_percent) / 100;
		} else {
			updatedNote.discount_amount = 0;
		}

		const taxable = subtotal - updatedNote.discount_amount;

		// Reset tax values
		updatedNote.cgst_rate = 0;
		updatedNote.cgst_amount = 0;
		updatedNote.sgst_rate = 0;
		updatedNote.sgst_amount = 0;
		updatedNote.igst_rate = 0;
		updatedNote.igst_amount = 0;

		const isSameState = updatedNote.customer_state_code === '33';
		const defaultTaxRate = 18;

		if (isSameState) {
			updatedNote.cgst_rate = defaultTaxRate / 2;
			updatedNote.sgst_rate = defaultTaxRate / 2;
			updatedNote.cgst_amount = (taxable * updatedNote.cgst_rate) / 100;
			updatedNote.sgst_amount = (taxable * updatedNote.sgst_rate) / 100;
		} else {
			updatedNote.igst_rate = defaultTaxRate;
			updatedNote.igst_amount = (taxable * updatedNote.igst_rate) / 100;
		}

		updatedNote.total_price = taxable + updatedNote.cgst_amount + updatedNote.sgst_amount + updatedNote.igst_amount;
		setFormDebitNote({ ...updatedNote });
	};

	const triggerPrint = () => {
		window.print();
	};

	const filteredInvoices = invoices.filter(inv => {
		const matchSearch = inv.customer_name.toLowerCase().includes(invoiceSearch.toLowerCase()) ||
			(inv.invoice_number && inv.invoice_number.toLowerCase().includes(invoiceSearch.toLowerCase()));
		const matchStatus = invoiceStatusFilter ? inv.status === invoiceStatusFilter : true;
		return matchSearch && matchStatus;
	});

	return (
		<>
			<div className="b2b-billing-container glass-card" style={{ padding: '32px', margin: '12px 0', borderRadius: '20px', background: 'var(--surface-color)', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-md)' }}>
				{/* Premium Layout Styles */}
				<style>{`
				@media print {
					:root, [data-theme="dark"], [data-theme="light"] {
						--text-primary: #000000 !important;
						--text-secondary: #333333 !important;
						--text-tertiary: #555555 !important;
						--border-color: #dddddd !important;
						--surface-color: #ffffff !important;
						--bg-color: #ffffff !important;
						--accent-color: #0d9488 !important;
					}
					.sidebar, .page-header, .no-print, button, .theme-toggle, .user-profile-menu {
						display: none !important;
					}
					html, body, #root, .app-container, .main-content, .b2b-billing-container {
						display: block !important;
						position: static !important;
						width: 100% !important;
						height: auto !important;
						min-height: auto !important;
						overflow: visible !important;
						margin: 0 !important;
						padding: 0 !important;
						background: white !important;
						color: black !important;
					}
					.print-invoice-area {
						display: block !important;
						position: relative !important;
						width: 100% !important;
						height: auto !important;
						background: white !important;
						color: black !important;
						padding: 40px !important;
						margin: 0 !important;
						box-shadow: none !important;
						border: none !important;
					}
				}
				.b2b-billing-container {
					animation: fadeInUp 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
				}
				.b2b-table-container {
					width: 100%;
					overflow-x: auto;
					overflow-y: visible;
					-webkit-overflow-scrolling: touch;
					margin-bottom: 1rem;
				}
				.b2b-input {
					background-color: var(--bg-input) !important;
					color: var(--text-primary) !important;
					border: 1px solid var(--border-color) !important;
					border-radius: 10px !important;
					padding: 0.65rem 0.9rem !important;
					font-size: 0.9rem !important;
					outline: none;
					transition: all 0.2s ease-in-out;
					width: 100%;
				}
				.b2b-input:focus {
					border-color: var(--accent-color) !important;
					background-color: var(--surface-color) !important;
					box-shadow: 0 0 0 3px var(--accent-subtle) !important;
				}
				.b2b-input::placeholder {
					color: var(--text-tertiary) !important;
				}
				.b2b-tooltip {
					position: relative;
					pointer-events: auto;
				}
				.b2b-tooltip::after {
					content: attr(data-tooltip);
					position: absolute;
					bottom: 125%;
					left: 50%;
					transform: translateX(-50%) translateY(4px);
					background: var(--bg-card, #0f172a);
					color: var(--text-primary, #f8fafc);
					padding: 6px 10px;
					border-radius: 6px;
					font-size: 11px;
					font-weight: 600;
					white-space: nowrap;
					opacity: 0;
					pointer-events: none;
					transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
					box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
					border: 1px solid var(--border-color, #334155);
					z-index: 99999;
				}
				.b2b-tooltip:hover::after {
					opacity: 1;
					transform: translateX(-50%) translateY(0);
				}
				.b2b-tooltip svg {
					pointer-events: none;
				}
				.b2b-btn {
					padding: 0.65rem 1.2rem;
					font-weight: 600;
					font-size: 0.9rem;
					border-radius: 10px;
					border: none;
					cursor: pointer;
					transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
					display: inline-flex;
					align-items: center;
					justify-content: center;
					gap: 6px;
				}
				.b2b-btn:hover {
					transform: translateY(-1px);
					box-shadow: var(--shadow-sm);
				}
				.b2b-btn-primary {
					background-color: var(--accent-color);
					color: white;
				}
				.b2b-btn-primary:hover {
					background-color: var(--accent-hover);
					box-shadow: 0 4px 12px var(--accent-subtle);
				}
				.b2b-btn-secondary {
					background-color: var(--bg-input);
					color: var(--text-primary);
					border: 1px solid var(--border-color);
				}
				.b2b-btn-secondary:hover {
					background-color: var(--bg-hover);
					border-color: var(--border-strong);
				}
				.b2b-btn-danger {
					background-color: var(--status-danger-bg);
					color: var(--status-danger);
					border: 1px solid rgba(239, 68, 68, 0.15);
				}
				.b2b-btn-danger:hover {
					background-color: var(--status-danger);
					color: white;
				}
				.b2b-btn-success {
					background-color: var(--status-active-bg);
					color: var(--status-active);
					border: 1px solid rgba(16, 185, 129, 0.15);
				}
				.b2b-btn-success:hover {
					background-color: var(--status-active);
					color: white;
				}
				.b2b-subtab-btn {
					background: transparent;
					border: 1px solid var(--border-color);
					color: var(--text-secondary);
					padding: 8px 18px;
					border-radius: 20px;
					font-weight: 600;
					font-size: 0.85rem;
					cursor: pointer;
					transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
				}
				.b2b-subtab-btn:hover {
					background: var(--bg-hover);
					border-color: var(--accent-color);
					color: var(--text-primary);
				}
				.b2b-subtab-btn.active {
					border-color: var(--accent-color);
					background: var(--accent-subtle);
					color: var(--accent-color);
					box-shadow: 0 4px 12px rgba(16, 185, 129, 0.08);
				}
				.b2b-form-section {
					background: var(--bg-hover);
					padding: 20px;
					border-radius: 12px;
					border: 1px solid var(--border-color);
					margin-bottom: 24px;
				}
				.b2b-form-section-title {
					font-size: 0.75rem;
					font-weight: 700;
					color: var(--text-secondary);
					text-transform: uppercase;
					letter-spacing: 0.05em;
					margin-bottom: 16px;
					border-bottom: 1px solid var(--border-color);
					padding-bottom: 8px;
				}
				.search-wrapper {
					position: relative;
					display: flex;
					align-items: center;
				}
				.search-icon {
					position: absolute;
					left: 12px;
					color: var(--text-tertiary);
					pointer-events: none;
				}
				.search-input {
					padding-left: 36px !important;
				}
				.clear-search-btn {
					position: absolute;
					right: 10px;
					background: transparent;
					border: none;
					color: var(--text-tertiary);
					cursor: pointer;
					display: flex;
					align-items: center;
					justify-content: center;
					padding: 4px;
					border-radius: 50%;
					transition: all 0.2s;
				}
				.clear-search-btn:hover {
					color: var(--text-primary);
					background: var(--bg-hover);
				}
				.form-label {
					font-size: 0.75rem;
					font-weight: 700;
					color: var(--text-secondary);
					text-transform: uppercase;
					letter-spacing: 0.05em;
					margin-bottom: 6px;
				}
				.b2b-table {
					width: 100%;
					border-collapse: separate;
					border-spacing: 0;
				}
				.b2b-table th {
					padding: 1.25rem 1.5rem;
					font-size: 0.75rem;
					font-weight: 800;
					color: var(--text-tertiary);
					text-transform: uppercase;
					letter-spacing: 1px;
					border-bottom: 1px solid var(--border-color);
					background: var(--bg-hover);
				}
				.b2b-table td {
					padding: 1.25rem 1.5rem;
					font-size: 0.9rem;
					border-bottom: 1px solid var(--border-color);
					transition: background-color 0.2s ease;
					color: var(--text-primary);
				}
				.b2b-table tr:last-child td {
					border-bottom: none;
				}
				.b2b-table tr:hover td {
					background-color: var(--bg-hover);
				}
				
				/* Redesigned Form UI Styles */
				.form-header-container {
					display: flex;
					justify-content: space-between;
					align-items: center;
					margin-bottom: 28px;
					border-bottom: 1px solid var(--border-color);
					padding-bottom: 20px;
				}
				.form-header-title {
					font-size: 1.6rem;
					font-weight: 800;
					color: var(--text-primary);
					margin: 0;
					letter-spacing: -0.025em;
					display: flex;
					align-items: center;
					gap: 12px;
				}
				.form-header-subtitle {
					font-size: 0.875rem;
					color: var(--text-secondary);
					margin: 4px 0 0 0;
				}
				.b2b-form-section {
					background: var(--surface-color) !important;
					border: 1px solid var(--border-color) !important;
					border-radius: 16px !important;
					padding: 24px !important;
					margin-bottom: 28px !important;
					box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.03), 0 2px 4px -1px rgba(0, 0, 0, 0.02) !important;
					transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
				}
				.b2b-form-section:hover {
					border-color: rgba(16, 185, 129, 0.3) !important;
					box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05), 0 4px 6px -2px rgba(0, 0, 0, 0.03) !important;
				}
				.b2b-form-section-title {
					font-size: 0.8rem !important;
					font-weight: 700 !important;
					color: var(--accent-color) !important;
					text-transform: uppercase !important;
					letter-spacing: 0.08em !important;
					margin-bottom: 20px !important;
					padding-bottom: 6px !important;
					border-bottom: 2px solid var(--accent-subtle) !important;
					display: inline-block !important;
				}
				.client-info-card {
					background: linear-gradient(135deg, var(--bg-hover) 0%, var(--surface-color) 100%) !important;
					border: 1px solid var(--border-color) !important;
					border-radius: 12px !important;
					padding: 18px 22px !important;
					display: flex;
					flex-direction: column;
					gap: 10px;
					position: relative;
					overflow: hidden;
					animation: fadeInScale 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
				}
				.client-info-card::before {
					content: '';
					position: absolute;
					left: 0;
					top: 0;
					bottom: 0;
					width: 4px;
					background: linear-gradient(180deg, var(--accent-color), #10b981);
				}
				.client-badge {
					display: inline-flex;
					align-items: center;
					padding: 4px 8px;
					background: var(--accent-subtle);
					color: var(--accent-color);
					border-radius: 6px;
					font-size: 0.75rem;
					font-weight: 700;
				}
				.items-table-header th {
					font-weight: 700 !important;
					text-transform: uppercase !important;
					font-size: 0.75rem !important;
					letter-spacing: 0.05em !important;
					color: var(--text-secondary) !important;
					border-bottom: 2px solid var(--border-color) !important;
					padding: 12px 16px !important;
					background: transparent !important;
				}
				.gst-select-input {
					padding: 0.65rem 1.5rem 0.65rem 0.75rem !important;
				}
				.items-table-row {
					transition: background-color 0.2s;
				}
				.items-table-row:hover {
					background-color: var(--bg-hover);
				}
				.delete-row-btn {
					background: transparent !important;
					border: none !important;
					color: var(--text-tertiary) !important;
					cursor: pointer;
					padding: 8px !important;
					border-radius: 50% !important;
					display: inline-flex !important;
					align-items: center;
					justify-content: center;
					transition: all 0.2s !important;
				}
				.delete-row-btn:hover {
					color: var(--status-danger) !important;
					background-color: var(--status-danger-bg) !important;
					transform: scale(1.1);
				}
				.add-row-btn {
					display: inline-flex;
					align-items: center;
					gap: 8px;
					padding: 8px 16px;
					font-weight: 600;
					font-size: 0.85rem;
					border-radius: 8px;
					border: 1px dashed var(--accent-color);
					background: transparent;
					color: var(--accent-color);
					cursor: pointer;
					transition: all 0.2s;
				}
				.add-row-btn:hover {
					background: var(--accent-subtle);
					border-style: solid;
					transform: translateY(-1px);
				}
				.summary-panel {
					background: linear-gradient(180deg, var(--bg-hover) 0%, var(--surface-color) 100%) !important;
					border: 1px solid var(--border-color) !important;
					border-radius: 16px !important;
					padding: 24px !important;
					box-shadow: var(--shadow-sm) !important;
				}
				.summary-row {
					display: flex;
					justify-content: space-between;
					align-items: center;
					margin-bottom: 14px;
					font-size: 0.9rem;
				}
				.summary-total-box {
					background: linear-gradient(135deg, var(--accent-subtle) 0%, rgba(16, 185, 129, 0.05) 100%) !important;
					border: 1px solid rgba(16, 185, 129, 0.15) !important;
					border-radius: 12px !important;
					padding: 16px !important;
					display: flex;
					justify-content: space-between;
					align-items: center;
					margin-top: 16px;
				}
				.days-input-left {
					border-top-right-radius: 0px !important;
					border-bottom-right-radius: 0px !important;
					border-right: none !important;
				}
				.days-label-right {
					background: var(--bg-hover) !important;
					border: 1px solid var(--border-color) !important;
					border-left: none !important;
					padding: 0.65rem 0.9rem !important;
					border-top-right-radius: 10px !important;
					border-bottom-right-radius: 10px !important;
					border-top-left-radius: 0px !important;
					border-bottom-left-radius: 0px !important;
					color: var(--text-secondary) !important;
					font-size: 0.9rem !important;
					white-space: nowrap !important;
					height: 42px !important;
					display: inline-flex !important;
					align-items: center !important;
					box-sizing: border-box !important;
				}
				@keyframes fadeInScale {
					from {
						opacity: 0;
						transform: scale(0.97);
					}
					to {
						opacity: 1;
						transform: scale(1);
					}
				}
			`}</style>

				{/* Sub Tabs */}
				{viewMode === 'list' && (
					<div className="sub-tabs-container no-print" style={{ display: 'flex', gap: '8px', marginBottom: '24px', borderBottom: '1px solid var(--border-color)', paddingBottom: '12px', flexWrap: 'wrap' }}>
						<button className={`b2b-subtab-btn ${activeSubTab === 'invoices' ? 'active' : ''}`} onClick={() => setActiveSubTab('invoices')}>Invoices</button>
						<button className={`b2b-subtab-btn ${activeSubTab === 'proformas' ? 'active' : ''}`} onClick={() => { setActiveSubTab('proformas'); setViewMode('list'); }}>Proforma Invoices</button>
						<button className={`b2b-subtab-btn ${activeSubTab === 'credit-notes' ? 'active' : ''}`} onClick={() => setActiveSubTab('credit-notes')}>Credit Notes</button>
						<button className={`b2b-subtab-btn ${activeSubTab === 'debit-notes' ? 'active' : ''}`} onClick={() => setActiveSubTab('debit-notes')}>Debit Notes</button>
						<button className={`b2b-subtab-btn ${activeSubTab === 'customers' ? 'active' : ''}`} onClick={() => setActiveSubTab('customers')}>B2B Clients</button>
						<button className={`b2b-subtab-btn ${activeSubTab === 'outstanding' ? 'active' : ''}`} onClick={() => setActiveSubTab('outstanding')}>Outstanding Report</button>
						<button className={`b2b-subtab-btn ${activeSubTab === 'locks' ? 'active' : ''}`} onClick={() => setActiveSubTab('locks')}>GST Filing Locks</button>
					</div>
				)}

				{/* LIST VIEW */}
				{viewMode === 'list' && activeSubTab === 'invoices' && (
					<div className="no-print">
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
							<div style={{ display: 'flex', gap: '10px' }}>
								<div className="search-wrapper" style={{ width: '240px', flex: 'none' }}>
									<svg className="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
										<circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
									</svg>
									<input
										type="text"
										placeholder="Search Invoices..."
										className="b2b-input search-input"
										value={invoiceSearch}
										onChange={(e) => setInvoiceSearch(e.target.value)}
									/>
									{invoiceSearch && (
										<button type="button" aria-label="Clear search" className="clear-search-btn" onClick={() => setInvoiceSearch('')}>✕</button>
									)}
								</div>
								<select
									value={invoiceStatusFilter}
									onChange={(e) => setInvoiceStatusFilter(e.target.value)}
									className="b2b-input"
									style={{ width: '160px' }}
								>
									<option value="">All Statuses</option>
									<option value="DRAFT">Draft</option>
									<option value="ISSUED">Issued</option>
									<option value="CANCELLED">Cancelled</option>
								</select>
							</div>
							{userRole === 'admin' && (
								<button
									onClick={() => {
										const today = new Date().toISOString().split('T')[0];
										setFormInvoice({
											invoice_date: today,
											customer_gstin: '',
											customer_name: '',
											customer_state: '',
											customer_state_code: '',
											customer_address: '',
											subtotal_price: 0,
											discount_percent: 0,
											discount_amount: 0,
											cgst_rate: 0,
											cgst_amount: 0,
											sgst_rate: 0,
											sgst_amount: 0,
											igst_rate: 0,
											igst_amount: 0,
											tds_tcs_type: 'NONE',
											tds_tcs_rate: 0,
											tds_tcs_amount: 0,
											transportation_charge: 0,
											total_price: 0,
											status: 'DRAFT',
											payment_status: 'UNPAID',
											paid_amount: 0,
											balance_amount: 0,
											customer_notes: DEFAULT_CUSTOMER_NOTES,
											items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019', gst_rate: 18 }]
										});
										fetchNextInvoiceNumber(today);
										setViewMode('create');
									}}
									className="b2b-btn b2b-btn-primary"
								>
									+ Create Invoice
								</button>
							)}
						</div>

						<div className="table-responsive" style={{ overflowX: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
							<table className="b2b-table">
								<thead>
									<tr>
										<th>Invoice#</th>
										<th>Client</th>
										<th>Date</th>
										<th>Total Amount</th>
										<th>GST Split</th>
										<th>Status</th>
										<th>Payment Status</th>
										<th style={{ textAlign: 'right' }}>Actions</th>
									</tr>
								</thead>
								<tbody>
									{filteredInvoices.map((inv) => (
										<tr key={inv.id}>
											<td>{inv.invoice_number || <span style={{ opacity: 0.5 }}>Draft</span>}</td>
											<td>
												<strong>{inv.customer_name}</strong>
												<div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '2px' }}>{inv.customer_gstin}</div>
											</td>
											<td>{inv.invoice_date ? inv.invoice_date.split('T')[0] : ''}</td>
											<td style={{ fontWeight: 'bold' }}>₹{inv.total_price.toFixed(2)}</td>
											<td style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
												{inv.cgst_amount > 0 && <div>CGST: ₹{inv.cgst_amount.toFixed(2)} ({(inv.cgst_rate)}%)</div>}
												{inv.sgst_amount > 0 && <div>SGST: ₹{inv.sgst_amount.toFixed(2)} ({(inv.sgst_rate)}%)</div>}
												{inv.igst_amount > 0 && <div>IGST: ₹{inv.igst_amount.toFixed(2)} ({(inv.igst_rate)}%)</div>}
											</td>
											<td>
												<span style={{
													padding: '6px 10px',
													borderRadius: '12px',
													fontSize: '11px',
													fontWeight: 600,
													background: inv.status === 'ISSUED' ? 'var(--status-active-bg)' : inv.status === 'CANCELLED' ? 'var(--status-danger-bg)' : 'var(--status-warning-bg)',
													color: inv.status === 'ISSUED' ? 'var(--status-active)' : inv.status === 'CANCELLED' ? 'var(--status-danger)' : 'var(--status-warning)'
												}}>
													{inv.status}
												</span>
												{inv.inventory_deducted && (
													<div style={{ fontSize: '10px', color: 'var(--status-active)', fontWeight: 600, marginTop: '4px' }}>
														✓ Stock Deducted
													</div>
												)}
											</td>
											<td>
												<span style={{
													padding: '6px 10px',
													borderRadius: '12px',
													fontSize: '11px',
													fontWeight: 600,
													background: inv.payment_status === 'PAID' ? 'var(--status-active-bg)' : inv.payment_status === 'PARTIAL' ? 'var(--status-warning-bg)' : 'var(--status-danger-bg)',
													color: inv.payment_status === 'PAID' ? 'var(--status-active)' : inv.payment_status === 'PARTIAL' ? 'var(--status-warning)' : 'var(--status-danger)'
												}}>
													{inv.payment_status}
												</span>
												<div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '4px' }}>Bal: ₹{inv.balance_amount.toFixed(2)}</div>
											</td>
											<td style={{ textAlign: 'right' }}>
												<div style={{ display: 'flex', gap: '6px', justifyContent: 'flex-end', flexWrap: 'nowrap' }}>
													{/* View Button */}
													<button
														onClick={() => {
															setSelectedInvoice(inv);
															setViewMode('preview');
														}}
														className="b2b-btn b2b-tooltip"
														data-tooltip="View Invoice"
														style={{
															width: '32px',
															height: '32px',
															borderRadius: '8px',
															display: 'inline-flex',
															alignItems: 'center',
															justifyContent: 'center',
															border: '1px solid var(--border-color)',
															background: 'var(--surface-color)',
															color: 'var(--text-secondary)',
															transition: 'all 0.2s',
															padding: 0
														}}
														onMouseEnter={(e) => {
															e.currentTarget.style.borderColor = 'var(--accent-color)';
															e.currentTarget.style.color = 'var(--accent-color)';
															e.currentTarget.style.background = 'var(--accent-subtle)';
														}}
														onMouseLeave={(e) => {
															e.currentTarget.style.borderColor = 'var(--border-color)';
															e.currentTarget.style.color = 'var(--text-secondary)';
															e.currentTarget.style.background = 'var(--surface-color)';
														}}
													>
														<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>
													</button>

													{inv.status === 'DRAFT' && userRole === 'admin' && (
														<>
															{/* Edit Button */}
															<button
																onClick={() => {
																	setFormInvoice({ ...inv });
																	setViewMode('edit');
																}}
																className="b2b-btn b2b-tooltip"
																data-tooltip="Edit Draft"
																style={{
																	width: '32px',
																	height: '32px',
																	borderRadius: '8px',
																	display: 'inline-flex',
																	alignItems: 'center',
																	justifyContent: 'center',
																	border: '1px solid var(--border-color)',
																	background: 'var(--surface-color)',
																	color: 'var(--text-secondary)',
																	transition: 'all 0.2s',
																	padding: 0
																}}
																onMouseEnter={(e) => {
																	e.currentTarget.style.borderColor = '#eab308'; // Amber
																	e.currentTarget.style.color = '#eab308';
																	e.currentTarget.style.background = 'rgba(234, 179, 8, 0.08)';
																}}
																onMouseLeave={(e) => {
																	e.currentTarget.style.borderColor = 'var(--border-color)';
																	e.currentTarget.style.color = 'var(--text-secondary)';
																	e.currentTarget.style.background = 'var(--surface-color)';
																}}
															>
																<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 1 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
															</button>

															{/* Issue Button */}
															<button
																onClick={() => handleIssueInvoice(inv.id!)}
																className="b2b-btn b2b-tooltip"
																data-tooltip="Issue / Activate Invoice"
																style={{
																	width: '32px',
																	height: '32px',
																	borderRadius: '8px',
																	display: 'inline-flex',
																	alignItems: 'center',
																	justifyContent: 'center',
																	border: '1px solid var(--border-color)',
																	background: 'var(--surface-color)',
																	color: 'var(--text-secondary)',
																	transition: 'all 0.2s',
																	padding: 0
																}}
																onMouseEnter={(e) => {
																	e.currentTarget.style.borderColor = 'var(--status-active)';
																	e.currentTarget.style.color = 'var(--status-active)';
																	e.currentTarget.style.background = 'var(--status-active-bg)';
																}}
																onMouseLeave={(e) => {
																	e.currentTarget.style.borderColor = 'var(--border-color)';
																	e.currentTarget.style.color = 'var(--text-secondary)';
																	e.currentTarget.style.background = 'var(--surface-color)';
																}}
															>
																<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="22 4 12 14.01 9 11.01"></polyline><rect x="2" y="2" width="20" height="20" rx="4"></rect></svg>
															</button>
														</>
													)}

													{inv.status === 'ISSUED' && userRole === 'admin' && (
														<>
															{/* Payment Button */}
															<button
																onClick={() => {
																	setPaymentInvoice(inv);
																	setPaymentAmount(inv.balance_amount);
																	setShowPaymentModal(true);
																}}
																className="b2b-btn b2b-tooltip"
																data-tooltip="Record Payment"
																style={{
																	width: '32px',
																	height: '32px',
																	borderRadius: '8px',
																	display: 'inline-flex',
																	alignItems: 'center',
																	justifyContent: 'center',
																	border: '1px solid var(--border-color)',
																	background: 'var(--surface-color)',
																	color: 'var(--text-secondary)',
																	transition: 'all 0.2s',
																	padding: 0
																}}
																onMouseEnter={(e) => {
																	e.currentTarget.style.borderColor = 'var(--status-active)';
																	e.currentTarget.style.color = 'var(--status-active)';
																	e.currentTarget.style.background = 'var(--status-active-bg)';
																}}
																onMouseLeave={(e) => {
																	e.currentTarget.style.borderColor = 'var(--border-color)';
																	e.currentTarget.style.color = 'var(--text-secondary)';
																	e.currentTarget.style.background = 'var(--surface-color)';
																}}
															>
																<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect><line x1="1" y1="10" x2="23" y2="10"></line></svg>
															</button>

															{/* Cancel Button */}
															<button
																onClick={() => handleCancelInvoice(inv.id!)}
																className="b2b-btn b2b-tooltip"
																data-tooltip="Cancel Invoice"
																style={{
																	width: '32px',
																	height: '32px',
																	borderRadius: '8px',
																	display: 'inline-flex',
																	alignItems: 'center',
																	justifyContent: 'center',
																	border: '1px solid var(--border-color)',
																	background: 'var(--surface-color)',
																	color: 'var(--text-secondary)',
																	transition: 'all 0.2s',
																	padding: 0
																}}
																onMouseEnter={(e) => {
																	e.currentTarget.style.borderColor = 'var(--status-danger)';
																	e.currentTarget.style.color = 'var(--status-danger)';
																	e.currentTarget.style.background = 'var(--status-danger-bg)';
																}}
																onMouseLeave={(e) => {
																	e.currentTarget.style.borderColor = 'var(--border-color)';
																	e.currentTarget.style.color = 'var(--text-secondary)';
																	e.currentTarget.style.background = 'var(--surface-color)';
																}}
															>
																<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"></line></svg>
															</button>

															{/* Deduct / Revert Inventory Toggle Button */}
															{inv.inventory_deducted ? (
																<button
																	onClick={() => handleRevertInventory(inv.id!)}
																	className="b2b-btn b2b-tooltip"
																	data-tooltip="Revert Stock Deduction"
																	style={{
																		width: '32px',
																		height: '32px',
																		borderRadius: '8px',
																		display: 'inline-flex',
																		alignItems: 'center',
																		justifyContent: 'center',
																		border: '1px solid var(--status-active)',
																		background: 'var(--status-active-bg)',
																		color: 'var(--status-active)',
																		transition: 'all 0.2s',
																		padding: 0
																	}}
																	onMouseEnter={(e) => {
																		e.currentTarget.style.borderColor = 'var(--status-danger)';
																		e.currentTarget.style.color = 'var(--status-danger)';
																		e.currentTarget.style.background = 'var(--status-danger-bg)';
																	}}
																	onMouseLeave={(e) => {
																		e.currentTarget.style.borderColor = 'var(--status-active)';
																		e.currentTarget.style.color = 'var(--status-active)';
																		e.currentTarget.style.background = 'var(--status-active-bg)';
																	}}
																>
																	<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="1 4 1 10 7 10"></polyline><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"></path></svg>
																</button>
															) : (
																<button
																	onClick={() => handleDeductInventory(inv.id!)}
																	className="b2b-btn b2b-tooltip"
																	data-tooltip="Deduct Inventory"
																	style={{
																		width: '32px',
																		height: '32px',
																		borderRadius: '8px',
																		display: 'inline-flex',
																		alignItems: 'center',
																		justifyContent: 'center',
																		border: '1px solid var(--border-color)',
																		background: 'var(--surface-color)',
																		color: 'var(--text-secondary)',
																		transition: 'all 0.2s',
																		padding: 0
																	}}
																	onMouseEnter={(e) => {
																		e.currentTarget.style.borderColor = '#eab308';
																		e.currentTarget.style.color = '#eab308';
																		e.currentTarget.style.background = 'rgba(234, 179, 8, 0.1)';
																	}}
																	onMouseLeave={(e) => {
																		e.currentTarget.style.borderColor = 'var(--border-color)';
																		e.currentTarget.style.color = 'var(--text-secondary)';
																		e.currentTarget.style.background = 'var(--surface-color)';
																	}}
																>
																	<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>
																</button>
															)}
														</>
													)}

													{/* Delete Button */}
													{userRole === 'admin' && (
														<button
															onClick={() => handleDeleteInvoice(inv.id!)}
															className="b2b-btn b2b-tooltip"
															data-tooltip={inv.status === 'DRAFT' ? "Delete Draft" : inv.status === 'CANCELLED' ? "Delete Cancelled Invoice" : "Delete Invoice (Warning: Tax Impact)"}
															style={{
																width: '32px',
																height: '32px',
																borderRadius: '8px',
																display: 'inline-flex',
																alignItems: 'center',
																justifyContent: 'center',
																border: '1px solid var(--border-color)',
																background: 'var(--surface-color)',
																color: 'var(--text-secondary)',
																transition: 'all 0.2s',
																padding: 0
															}}
															onMouseEnter={(e) => {
																e.currentTarget.style.borderColor = 'var(--status-danger)';
																e.currentTarget.style.color = 'var(--status-danger)';
																e.currentTarget.style.background = 'var(--status-danger-bg)';
															}}
															onMouseLeave={(e) => {
																e.currentTarget.style.borderColor = 'var(--border-color)';
																e.currentTarget.style.color = 'var(--text-secondary)';
																e.currentTarget.style.background = 'var(--surface-color)';
															}}
														>
															<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
														</button>
													)}
												</div>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					</div>
				)}



				{/* PROFORMA INVOICES VIEW */}
				{viewMode === 'list' && activeSubTab === 'proformas' && (
					<div className="no-print" style={{ animation: 'fadeInScale 0.2s ease-out' }}>
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px', gap: '12px', flexWrap: 'wrap' }}>
							<div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap', alignItems: 'center' }}>
								<div className="search-wrapper" style={{ width: '260px', flex: 'none' }}>
									<svg className="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
										<circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
									</svg>
									<input
										type="text"
										placeholder="Search Proformas..."
										className="b2b-input search-input"
										value={proformaSearch}
										onChange={(e) => setProformaSearch(e.target.value)}
									/>
									{proformaSearch && (
										<button type="button" aria-label="Clear search" className="clear-search-btn" onClick={() => setProformaSearch('')}>✕</button>
									)}
								</div>
								<select
									className="b2b-input"
									style={{ width: '160px', height: '42px' }}
									value={proformaStatusFilter}
									onChange={(e) => setProformaStatusFilter(e.target.value)}
								>
									<option value="">All Statuses</option>
									<option value="DRAFT">Draft</option>
									<option value="SENT">Sent</option>
									<option value="ACCEPTED">Accepted</option>
									<option value="CONVERTED_TO_INVOICE">Converted</option>
									<option value="REJECTED">Rejected</option>
									<option value="EXPIRED">Expired</option>
									<option value="CANCELLED">Cancelled</option>
								</select>
								<button
									className="b2b-btn b2b-btn-secondary"
									onClick={handleCheckExpiredProformas}
									style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', height: '42px' }}
								>
									🕒 Check Expiry
								</button>
							</div>

							{userRole === 'admin' && (
								<button
									className="b2b-btn b2b-btn-primary"
									onClick={() => {
										setFormProforma({
											note_date: new Date().toISOString().split('T')[0],
											valid_until: new Date(Date.now() + 30*24*60*60*1000).toISOString().split('T')[0], // 30 days valid by default
											customer_gstin: '',
											customer_name: '',
											customer_state: '',
											customer_state_code: '',
											customer_address: '',
											subtotal_price: 0,
											discount_percent: 0,
											discount_amount: 0,
											cgst_rate: 0,
											cgst_amount: 0,
											sgst_rate: 0,
											sgst_amount: 0,
											igst_rate: 0,
											igst_amount: 0,
											total_price: 0,
											advance_paid: 0,
											status: 'DRAFT',
											revision_number: 1,
											items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
										});
										setViewMode('create-pf');
										fetchNextProformaNumber();
									}}
									style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}
								>
									+ Create Proforma
								</button>
							)}
						</div>

						<div className="b2b-table-container">
							<table className="b2b-table">
								<thead>
									<tr>
										<th>Proforma No</th>
										<th>Date</th>
										<th>Customer</th>
										<th style={{ textAlign: 'right' }}>Total (₹)</th>
										<th style={{ textAlign: 'right' }}>Advance (₹)</th>
										<th style={{ textAlign: 'center' }}>Rev</th>
										<th>Valid Until</th>
										<th>Status</th>
										<th style={{ textAlign: 'right' }} className="no-print">Actions</th>
									</tr>
								</thead>
								<tbody>
									{proformas
										.filter(pf => {
											const matchesSearch = pf.proforma_number?.toLowerCase().includes(proformaSearch.toLowerCase()) ||
												pf.customer_name?.toLowerCase().includes(proformaSearch.toLowerCase()) ||
												pf.customer_gstin?.toLowerCase().includes(proformaSearch.toLowerCase());
											const matchesStatus = proformaStatusFilter === '' || pf.status === proformaStatusFilter;
											return matchesSearch && matchesStatus;
										})
										.map(pf => (
											<tr key={pf.id}>
												<td style={{ fontWeight: '600' }}>
													{pf.proforma_number || 'DRAFT'}
												</td>
												<td>{new Date(pf.note_date).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })}</td>
												<td>
													<div style={{ fontWeight: '500' }}>{pf.customer_name}</div>
													<div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{pf.customer_gstin}</div>
												</td>
												<td style={{ textAlign: 'right', fontWeight: '600' }}>
													₹{pf.total_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}
												</td>
												<td style={{ textAlign: 'right', color: pf.advance_paid > 0 ? 'var(--success-color)' : 'var(--text-secondary)' }}>
													₹{pf.advance_paid.toLocaleString('en-IN', { minimumFractionDigits: 2 })}
												</td>
												<td style={{ textAlign: 'center' }}>
													<span style={{ background: 'var(--bg-hover)', padding: '2px 6px', borderRadius: '4px', fontSize: '0.8rem' }}>
														v{pf.revision_number}
													</span>
												</td>
												<td>
													{pf.valid_until ? (
														<span style={{ color: new Date(pf.valid_until) < new Date() && pf.status === 'SENT' ? 'var(--danger-color)' : 'inherit' }}>
															{new Date(pf.valid_until).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })}
														</span>
													) : '—'}
												</td>
												<td>
													{(() => {
														const { bg, color, border } = getProformaStatusStyle(pf.status);
														return (
															<span style={{ 
																display: 'inline-flex',
																alignItems: 'center',
																padding: '6px 12px',
																borderRadius: '9999px',
																fontSize: '0.72rem',
																fontWeight: 700,
																backgroundColor: bg,
																color: color,
																border: `1px solid ${border}`,
																textTransform: 'uppercase',
																letterSpacing: '0.03em',
																gap: '6px'
															}}>
																<span style={{ width: '6px', height: '6px', borderRadius: '50%', backgroundColor: color }} />
																{pf.status === 'CONVERTED_TO_INVOICE' ? 'CONVERTED' : pf.status.replace(/_/g, ' ')}
															</span>
														);
													})()}
												</td>
												<td className="no-print" style={{ width: '1%', whiteSpace: 'nowrap' }}>
													<div style={{ display: 'flex', gap: '6px', justifyContent: 'flex-end', flexWrap: 'nowrap', whiteSpace: 'nowrap' }}>
														<button
															className="b2b-btn b2b-btn-secondary b2b-tooltip"
															onClick={() => {
																setSelectedProforma(pf);
																setViewMode('preview-pf');
															}}
															data-tooltip="View Proforma"
															style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
															onMouseEnter={(e) => { e.currentTarget.style.borderColor = '#9ca3af'; e.currentTarget.style.background = '#f9fafb'; }}
															onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; }}
														>
															<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
																<circle cx="12" cy="12" r="3"></circle>
															</svg>
														</button>

														{userRole === 'admin' && (
															<button
																className="b2b-btn b2b-btn-danger b2b-tooltip"
																onClick={() => handleDeleteProforma(pf.id)}
																data-tooltip="Delete Proforma"
																style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
																onMouseEnter={(e) => { e.currentTarget.style.borderColor = 'var(--status-danger)'; e.currentTarget.style.background = 'var(--status-danger-bg)'; e.currentTarget.style.color = 'var(--status-danger)'; }}
																onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; e.currentTarget.style.color = '#374151'; }}
															>
																<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																	<polyline points="3 6 5 6 21 6"></polyline>
																	<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
																</svg>
															</button>
														)}

														{pf.status === 'DRAFT' && userRole === 'admin' && (
															<>
																<button
																	className="b2b-btn b2b-btn-secondary b2b-tooltip"
																	onClick={() => {
																		setFormProforma(pf);
																		setViewMode('edit-pf');
																		fetchNextProformaNumber(pf.note_date);
																	}}
																	data-tooltip="Edit Proforma"
																	style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
																	onMouseEnter={(e) => { e.currentTarget.style.borderColor = '#9ca3af'; e.currentTarget.style.background = '#f9fafb'; }}
																	onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; }}
																>
																	<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																		<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
																		<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
																	</svg>
																</button>
																<button
																	className="b2b-btn b2b-btn-primary b2b-tooltip"
																	onClick={() => handleIssueProforma(pf.id)}
																	data-tooltip="Send Proforma"
																	style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
																	onMouseEnter={(e) => { e.currentTarget.style.borderColor = '#9ca3af'; e.currentTarget.style.background = '#f9fafb'; }}
																	onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; }}
																>
																	<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																		<line x1="22" y1="2" x2="11" y2="13"></line>
																		<polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
																	</svg>
																</button>
															</>
														)}

														{pf.status === 'SENT' && userRole === 'admin' && (
															<>
																<button
																	className="b2b-btn b2b-btn-secondary b2b-tooltip"
																	onClick={() => handleAcceptProforma(pf.id)}
																	data-tooltip="Accept Proforma"
																	style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
																	onMouseEnter={(e) => { e.currentTarget.style.borderColor = 'var(--status-success)'; e.currentTarget.style.background = 'var(--status-success-bg)'; e.currentTarget.style.color = 'var(--status-success)'; }}
																	onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; e.currentTarget.style.color = '#374151'; }}
																>
																	<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
																		<polyline points="20 6 9 17 4 12"></polyline>
																	</svg>
																</button>
																<button
																	className="b2b-btn b2b-btn-secondary b2b-tooltip"
																	onClick={() => handleCreateRevision(pf.id)}
																	data-tooltip="Revise Proforma"
																	style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
																	onMouseEnter={(e) => { e.currentTarget.style.borderColor = '#eab308'; e.currentTarget.style.background = 'rgba(234, 179, 8, 0.1)'; e.currentTarget.style.color = '#eab308'; }}
																	onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; e.currentTarget.style.color = '#374151'; }}
																>
																	<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																		<path d="M12 20h9"></path>
																		<path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
																	</svg>
																</button>
																<button
																	className="b2b-btn b2b-btn-danger b2b-tooltip"
																	onClick={() => handleCancelProforma(pf.id)}
																	data-tooltip="Cancel Proforma"
																	style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
																	onMouseEnter={(e) => { e.currentTarget.style.borderColor = 'var(--status-danger)'; e.currentTarget.style.background = 'var(--status-danger-bg)'; e.currentTarget.style.color = 'var(--status-danger)'; }}
																	onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; e.currentTarget.style.color = '#374151'; }}
																>
																	<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																		<circle cx="12" cy="12" r="10"></circle>
																		<line x1="4.93" y1="4.93" x2="19.07" y2="19.07"></line>
																	</svg>
																</button>
															</>
														)}

														{(pf.status === 'ACCEPTED' || pf.status === 'SENT') && userRole === 'admin' && (
															<button
																className="b2b-btn b2b-btn-primary b2b-tooltip"
																onClick={() => handleConvertToTaxInvoice(pf.id)}
																data-tooltip="Convert to Tax Invoice"
																style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
																onMouseEnter={(e) => { e.currentTarget.style.borderColor = 'var(--accent-color)'; e.currentTarget.style.background = 'var(--accent-subtle)'; e.currentTarget.style.color = 'var(--accent-color)'; }}
																onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; e.currentTarget.style.color = '#374151'; }}
															>
																<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																	<rect x="1" y="4" width="22" height="16" rx="2" ry="2"></rect>
																	<line x1="1" y1="10" x2="23" y2="10"></line>
																</svg>
															</button>
														)}

														{pf.status === 'REJECTED' && userRole === 'admin' && (
															<button
																className="b2b-btn b2b-btn-secondary b2b-tooltip"
																onClick={() => handleCreateRevision(pf.id)}
																data-tooltip="Revise Proforma"
																style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '32px', height: '32px', padding: '0', minHeight: 'auto', borderRadius: '12px', background: '#ffffff', color: '#374151', border: '1px solid #e5e7eb', cursor: 'pointer', transition: 'all 0.2s' }}
																onMouseEnter={(e) => { e.currentTarget.style.borderColor = '#eab308'; e.currentTarget.style.background = 'rgba(234, 179, 8, 0.1)'; e.currentTarget.style.color = '#eab308'; }}
																onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e5e7eb'; e.currentTarget.style.background = '#ffffff'; e.currentTarget.style.color = '#374151'; }}
															>
																<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																	<path d="M12 20h9"></path>
																	<path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path>
																</svg>
															</button>
														)}
													</div>
												</td>
											</tr>
										))}
									{proformas.length === 0 && (
										<tr>
											<td colSpan={9} style={{ textAlign: 'center', padding: '24px', color: 'var(--text-secondary)' }}>
												No proforma invoices found. Click "Create Proforma" to start.
											</td>
										</tr>
									)}
								</tbody>
							</table>
						</div>
					</div>
				)}

				{/* CUSTOMER REGISTRY VIEW */}
				{viewMode === 'list' && activeSubTab === 'customers' && (
					<div>
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
							<div className="search-wrapper" style={{ width: '280px', flex: 'none' }}>
								<svg className="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
									<circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
								</svg>
								<input
									type="text"
									placeholder="Search Clients..."
									className="b2b-input search-input"
									value={customerSearch}
									onChange={(e) => setCustomerSearch(e.target.value)}
								/>
								{customerSearch && (
									<button type="button" aria-label="Clear search" className="clear-search-btn" onClick={() => setCustomerSearch('')}>✕</button>
								)}
							</div>
							{userRole === 'admin' && (
								<button
									onClick={() => {
										setEditingCustomer({
											legal_name: '',
											gstin: '',
											billing_address: '',
											shipping_address: '',
											state: '',
											state_code: ''
										});
										setSameAsBilling(true);
										setShowCustomerModal(true);
									}}
									className="b2b-btn b2b-btn-primary"
								>
									+ Add Client
								</button>
							)}
						</div>

						<div className="table-responsive" style={{ overflowX: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
							<table className="b2b-table">
								<thead>
									<tr>
										<th>Legal Name</th>
										<th>GSTIN</th>
										<th>PAN</th>
										<th>State (Code)</th>
										<th>Billing Address</th>
										<th style={{ textAlign: 'right' }}>Actions</th>
									</tr>
								</thead>
								<tbody>
									{customers.filter(c => c.legal_name.toLowerCase().includes(customerSearch.toLowerCase()) || c.gstin.toLowerCase().includes(customerSearch.toLowerCase())).map((cust) => (
										<tr key={cust.id}>
											<td><strong>{cust.legal_name}</strong> {cust.trade_name && <span style={{ display: 'block', fontSize: '11px', opacity: 0.7 }}>Trade: {cust.trade_name}</span>}</td>
											<td>{cust.gstin}</td>
											<td>{cust.pan}</td>
											<td>{cust.state} ({cust.state_code})</td>
											<td style={{ maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{cust.billing_address}</td>
											<td style={{ textAlign: 'right' }}>
												<div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
													<button
														onClick={() => {
															setLedgerCustomer(cust);
															loadCustomerLedger(cust.id!);
															setShowLedgerModal(true);
														}}
														className="b2b-btn b2b-btn-secondary"
														style={{ padding: '6px', minHeight: 'auto', borderRadius: '8px' }}
														title="View Ledger Statement"
													>
														<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
															<rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
															<line x1="9" y1="9" x2="15" y2="9" />
															<line x1="9" y1="13" x2="15" y2="13" />
															<line x1="9" y1="17" x2="15" y2="17" />
														</svg>
													</button>
													<button
														onClick={() => {
															setEditingCustomer({ ...cust });
															setSameAsBilling(!cust.shipping_address || cust.shipping_address === cust.billing_address);
															setShowCustomerModal(true);
														}}
														className="b2b-btn b2b-btn-secondary"
														style={{ padding: '6px', minHeight: 'auto', borderRadius: '8px' }}
														title="Edit Client"
													>
														<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
															<path d="M12 20h9" /><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
														</svg>
													</button>
													<button
														onClick={() => handleDeleteCustomer(cust.id!)}
														className="b2b-btn b2b-btn-danger"
														style={{ padding: '6px', minHeight: 'auto', borderRadius: '8px' }}
														title="Delete Client"
													>
														<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
															<polyline points="3 6 5 6 21 6" /><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /><line x1="10" y1="11" x2="10" y2="17" /><line x1="14" y1="11" x2="14" y2="17" />
														</svg>
													</button>
												</div>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					</div>
				)}

				{/* CREDIT NOTES VIEW */}
				{viewMode === 'list' && activeSubTab === 'credit-notes' && (
					<div className="no-print">
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
							<div style={{ display: 'flex', gap: '10px' }}>
								<div className="search-wrapper" style={{ width: '240px', flex: 'none' }}>
									<svg className="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
										<circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
									</svg>
									<input
										type="text"
										placeholder="Search Credit Notes..."
										className="b2b-input search-input"
										value={creditSearch}
										onChange={(e) => setCreditSearch(e.target.value)}
									/>
									{creditSearch && (
										<button type="button" aria-label="Clear search" className="clear-search-btn" onClick={() => setCreditSearch('')}>✕</button>
									)}
								</div>
							</div>
							{userRole === 'admin' && (
								<button
									onClick={() => {
										setFormCreditNote({
											note_date: new Date().toISOString().split('T')[0],
											customer_gstin: '',
											customer_name: '',
											customer_state: '',
											customer_state_code: '',
											customer_address: '',
											subtotal_price: 0,
											discount_percent: 0,
											discount_amount: 0,
											cgst_rate: 0,
											cgst_amount: 0,
											sgst_rate: 0,
											sgst_amount: 0,
											igst_rate: 0,
											igst_amount: 0,
											total_price: 0,
											status: 'DRAFT',
											reason: '',
											invoice_id: undefined,
											items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
										});
										setViewMode('create-cn');
									}}
									className="b2b-btn b2b-btn-primary"
								>
									+ Create Credit Note
								</button>
							)}
						</div>

						<div className="table-responsive" style={{ overflowX: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
							<table className="b2b-table">
								<thead>
									<tr>
										<th>Note#</th>
										<th>Linked Invoice</th>
										<th>Client</th>
										<th>Date</th>
										<th>Total Amount</th>
										<th>Reason</th>
										<th>Status</th>
										<th style={{ textAlign: 'right' }}>Actions</th>
									</tr>
								</thead>
								<tbody>
									{creditNotes.filter(cn => 
										cn.customer_name.toLowerCase().includes(creditSearch.toLowerCase()) ||
										(cn.credit_note_number && cn.credit_note_number.toLowerCase().includes(creditSearch.toLowerCase()))
									).map((cn) => (
										<tr key={cn.id}>
											<td>{cn.credit_note_number || <span style={{ opacity: 0.5 }}>Draft</span>}</td>
											<td>{cn.invoice_number || 'Unlinked'}</td>
											<td>
												<strong>{cn.customer_name}</strong>
												<div style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>{cn.customer_gstin}</div>
											</td>
											<td>{cn.note_date ? cn.note_date.split('T')[0] : ''}</td>
											<td style={{ fontWeight: 'bold' }}>₹{cn.total_price.toFixed(2)}</td>
											<td>{cn.reason || <span style={{ opacity: 0.5 }}>N/A</span>}</td>
											<td>
												<span style={{
													padding: '6px 10px',
													borderRadius: '12px',
													fontSize: '11px',
													fontWeight: 600,
													background: cn.status === 'ISSUED' ? 'var(--status-active-bg)' : cn.status === 'CANCELLED' ? 'var(--status-danger-bg)' : 'var(--status-warning-bg)',
													color: cn.status === 'ISSUED' ? 'var(--status-active)' : cn.status === 'CANCELLED' ? 'var(--status-danger)' : 'var(--status-warning)'
												}}>
													{cn.status}
												</span>
											</td>
											<td style={{ textAlign: 'right' }}>
												<div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
													<button
														onClick={() => {
															setSelectedCreditNote(cn);
															setViewMode('preview-cn');
														}}
														className="b2b-btn b2b-btn-secondary"
														style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
													>
														View
													</button>
													{cn.status === 'DRAFT' && userRole === 'admin' && (
														<>
															<button
																onClick={() => {
																	setFormCreditNote({ ...cn, note_date: cn.note_date.split('T')[0] });
																	setViewMode('edit-cn');
																}}
																className="b2b-btn b2b-btn-secondary"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Edit
															</button>
															<button
																onClick={() => handleIssueCreditNote(cn.id)}
																className="b2b-btn b2b-btn-success"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Issue
															</button>
															<button
																onClick={() => handleDeleteCreditNote(cn.id)}
																className="b2b-btn b2b-btn-danger"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Delete
															</button>
														</>
													)}
													{cn.status === 'ISSUED' && userRole === 'admin' && (
														<button
															onClick={() => handleCancelCreditNote(cn.id)}
															className="b2b-btn b2b-btn-danger"
															style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
														>
															Cancel
														</button>
													)}
												</div>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					</div>
				)}

				{/* DEBIT NOTES VIEW */}
				{viewMode === 'list' && activeSubTab === 'debit-notes' && (
					<div className="no-print">
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
							<div style={{ display: 'flex', gap: '10px' }}>
								<div className="search-wrapper" style={{ width: '240px', flex: 'none' }}>
									<svg className="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
										<circle cx="11" cy="11" r="8" /><path d="m21 21-4.3-4.3" />
									</svg>
									<input
										type="text"
										placeholder="Search Debit Notes..."
										className="b2b-input search-input"
										value={debitSearch}
										onChange={(e) => setDebitSearch(e.target.value)}
									/>
									{debitSearch && (
										<button type="button" aria-label="Clear search" className="clear-search-btn" onClick={() => setDebitSearch('')}>✕</button>
									)}
								</div>
							</div>
							{userRole === 'admin' && (
								<button
									onClick={() => {
										setFormDebitNote({
											note_date: new Date().toISOString().split('T')[0],
											customer_gstin: '',
											customer_name: '',
											customer_state: '',
											customer_state_code: '',
											customer_address: '',
											subtotal_price: 0,
											discount_percent: 0,
											discount_amount: 0,
											cgst_rate: 0,
											cgst_amount: 0,
											sgst_rate: 0,
											sgst_amount: 0,
											igst_rate: 0,
											igst_amount: 0,
											total_price: 0,
											status: 'DRAFT',
											reason: '',
											invoice_id: undefined,
											items: [{ item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }]
										});
										setViewMode('create-dn');
									}}
									className="b2b-btn b2b-btn-primary"
								>
									+ Create Debit Note
								</button>
							)}
						</div>

						<div className="table-responsive" style={{ overflowX: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
							<table className="b2b-table">
								<thead>
									<tr>
										<th>Note#</th>
										<th>Linked Invoice</th>
										<th>Client</th>
										<th>Date</th>
										<th>Total Amount</th>
										<th>Reason</th>
										<th>Status</th>
										<th style={{ textAlign: 'right' }}>Actions</th>
									</tr>
								</thead>
								<tbody>
									{debitNotes.filter(dn => 
										dn.customer_name.toLowerCase().includes(debitSearch.toLowerCase()) ||
										(dn.debit_note_number && dn.debit_note_number.toLowerCase().includes(debitSearch.toLowerCase()))
									).map((dn) => (
										<tr key={dn.id}>
											<td>{dn.debit_note_number || <span style={{ opacity: 0.5 }}>Draft</span>}</td>
											<td>{dn.invoice_number || 'Unlinked'}</td>
											<td>
												<strong>{dn.customer_name}</strong>
												<div style={{ fontSize: '11px', color: 'var(--text-tertiary)' }}>{dn.customer_gstin}</div>
											</td>
											<td>{dn.note_date ? dn.note_date.split('T')[0] : ''}</td>
											<td style={{ fontWeight: 'bold' }}>₹{dn.total_price.toFixed(2)}</td>
											<td>{dn.reason || <span style={{ opacity: 0.5 }}>N/A</span>}</td>
											<td>
												<span style={{
													padding: '6px 10px',
													borderRadius: '12px',
													fontSize: '11px',
													fontWeight: 600,
													background: dn.status === 'ISSUED' ? 'var(--status-active-bg)' : dn.status === 'CANCELLED' ? 'var(--status-danger-bg)' : 'var(--status-warning-bg)',
													color: dn.status === 'ISSUED' ? 'var(--status-active)' : dn.status === 'CANCELLED' ? 'var(--status-danger)' : 'var(--status-warning)'
												}}>
													{dn.status}
												</span>
											</td>
											<td style={{ textAlign: 'right' }}>
												<div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
													<button
														onClick={() => {
															setSelectedDebitNote(dn);
															setViewMode('preview-dn');
														}}
														className="b2b-btn b2b-btn-secondary"
														style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
													>
														View
													</button>
													{dn.status === 'DRAFT' && userRole === 'admin' && (
														<>
															<button
																onClick={() => {
																	setFormDebitNote({ ...dn, note_date: dn.note_date.split('T')[0] });
																	setViewMode('edit-dn');
																}}
																className="b2b-btn b2b-btn-secondary"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Edit
															</button>
															<button
																onClick={() => handleIssueDebitNote(dn.id)}
																className="b2b-btn b2b-btn-success"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Issue
															</button>
															<button
																onClick={() => handleDeleteDebitNote(dn.id)}
																className="b2b-btn b2b-btn-danger"
																style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
															>
																Delete
															</button>
														</>
													)}
													{dn.status === 'ISSUED' && userRole === 'admin' && (
														<button
															onClick={() => handleCancelDebitNote(dn.id)}
															className="b2b-btn b2b-btn-danger"
															style={{ padding: '4px 10px', fontSize: '0.8rem', minHeight: 'auto' }}
														>
															Cancel
														</button>
													)}
												</div>
											</td>
										</tr>
									))}
								</tbody>
							</table>
						</div>
					</div>
				)}

				{/* OUTSTANDING REPORT VIEW */}
				{viewMode === 'list' && activeSubTab === 'outstanding' && (
					<div>
						<div style={{ marginBottom: '24px', background: 'linear-gradient(135deg, rgba(16,185,129,0.05) 0%, rgba(99,102,241,0.05) 100%)', padding: '24px', borderRadius: '16px', border: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
							<div>
								<h3 style={{ margin: 0, fontWeight: 800, fontSize: '1.4rem' }}>Outstanding Overdue Aging</h3>
								<p style={{ margin: '4px 0 0 0', color: 'var(--text-secondary)', fontSize: '0.9rem' }}>Real-time aging analysis of customer receivables grouped by overdue duration.</p>
							</div>
							<div style={{ background: 'var(--surface-color)', padding: '12px 20px', borderRadius: '12px', border: '1px solid var(--border-color)', textAlign: 'right' }}>
								<div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Total Receivables</div>
								<div style={{ fontSize: '1.6rem', fontWeight: 800, color: 'var(--accent-color)', marginTop: '2px' }}>
									₹{outstandingReport.reduce((acc, curr) => acc + curr.days_0_30 + curr.days_31_60 + curr.days_60_plus, 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
								</div>
							</div>
						</div>

						<div className="table-responsive" style={{ overflowX: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)', boxShadow: 'var(--shadow-sm)' }}>
							<table className="b2b-table">
								<thead>
									<tr>
										<th>Client Name</th>
										<th>GSTIN</th>
										<th>0 - 30 Days</th>
										<th>31 - 60 Days</th>
										<th>61+ Days</th>
										<th style={{ fontWeight: 'bold' }}>Total Outstanding</th>
									</tr>
								</thead>
								<tbody>
									{outstandingReport.length === 0 ? (
										<tr>
											<td colSpan={6} style={{ textAlign: 'center', padding: '40px', color: 'var(--text-tertiary)' }}>No outstanding client balances found.</td>
										</tr>
									) : (
										outstandingReport.map((out) => (
											<tr key={out.customer_id}>
												<td><strong>{out.customer_name}</strong></td>
												<td>{out.gstin}</td>
												<td style={{ color: out.days_0_30 > 0 ? 'var(--text-primary)' : 'var(--text-tertiary)' }}>₹{out.days_0_30.toFixed(2)}</td>
												<td style={{ color: out.days_31_60 > 0 ? 'var(--status-warning)' : 'var(--text-tertiary)' }}>₹{out.days_31_60.toFixed(2)}</td>
												<td style={{ color: out.days_60_plus > 0 ? 'var(--status-danger)' : 'var(--text-tertiary)' }}>₹{out.days_60_plus.toFixed(2)}</td>
												<td style={{ fontWeight: 'bold', color: 'var(--accent-color)' }}>₹{out.total_due.toFixed(2)}</td>
											</tr>
										))
									)}
								</tbody>
							</table>
						</div>
					</div>
				)}

				{/* GST PERIOD LOCKS VIEW */}
				{viewMode === 'list' && activeSubTab === 'locks' && (
					<div>
						<div style={{ marginBottom: '24px', background: 'var(--bg-hover)', padding: '20px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
							<h4 style={{ margin: 0, fontWeight: 700 }}>GST Return Lock Policy</h4>
							<p style={{ margin: '6px 0 0 0', color: 'var(--text-secondary)', fontSize: '0.85rem', lineHeight: 1.5 }}>
								Locking a filing period prevents historical invoice value modifications, additions, and deletions for that month. Payments and credit notes can still be posted against existing bills inside locked periods.
							</p>
						</div>

						<div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '20px' }}>
							{Array.from({ length: 12 }).map((_, idx) => {
								const d = new Date();
								d.setMonth(d.getMonth() - idx);
								const m = d.getMonth() + 1;
								const y = d.getFullYear();
								const period = gstPeriods.find(p => p.month === m && p.year === y);
								const isLocked = period ? period.status === 'LOCKED' : false;
								const monthName = d.toLocaleString('default', { month: 'long' });

								return (
									<div key={idx} style={{ 
										background: 'var(--surface-color)', 
										border: '1px solid var(--border-color)', 
										borderRadius: '16px', 
										padding: '20px', 
										display: 'flex', 
										justifyContent: 'space-between', 
										alignItems: 'center',
										boxShadow: 'var(--shadow-sm)'
									}}>
										<div>
											<strong style={{ fontSize: '1.05rem' }}>{monthName} {y}</strong>
											<div style={{ fontSize: '0.8rem', color: isLocked ? 'var(--status-danger)' : 'var(--status-active)', fontWeight: 600, marginTop: '4px' }}>
												{isLocked ? '🔒 FILING LOCKED' : '🔓 OPEN FOR EDITS'}
											</div>
										</div>
										{userRole === 'admin' && (
											<button 
												onClick={() => toggleGSTPeriod(m, y, isLocked ? 'LOCKED' : 'OPEN')}
												className={`b2b-btn ${isLocked ? 'b2b-btn-secondary' : 'b2b-btn-danger'}`}
												style={{ padding: '8px 14px', fontSize: '0.8rem' }}
											>
												{isLocked ? 'Unlock' : 'Lock Filing'}
											</button>
										)}
									</div>
								);
							})}
						</div>
					</div>
				)}

				{/* BILL CREATOR & EDITOR VIEW */}
				{(viewMode === 'create' || viewMode === 'edit') && (
					<div className="b2b-form-area no-print" style={{ color: 'var(--text-primary)' }}>
						<div className="form-header-container">
							<div>
								<h3 className="form-header-title">
									<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--accent-color)' }}>
										<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
										<polyline points="14 2 14 8 20 8" />
										<line x1="16" y1="13" x2="8" y2="13" />
										<line x1="16" y1="17" x2="8" y2="17" />
										<polyline points="10 9 9 9 8 9" />
									</svg>
									{viewMode === 'create' ? 'Create B2B Invoice' : 'Edit B2B Invoice'}
								</h3>
								<p className="form-header-subtitle">Fill in the details below to generate a GST-compliant invoice.</p>
							</div>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary" style={{ height: '40px' }}>
								&larr; Back to List
							</button>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Client Information</div>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', alignItems: 'start' }}>
								{/* Customer Selector */}
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Select B2B Customer*</label>
									<select
										className="b2b-input"
										value={formInvoice.customer_id || ''}
										onChange={(e) => {
											const cId = Number(e.target.value);
											const selected = customers.find(c => c.id === cId);
											if (selected) {
												const updated = {
													...formInvoice,
													customer_id: selected.id,
													customer_gstin: selected.gstin,
													customer_name: selected.legal_name,
													customer_state: selected.state,
													customer_state_code: selected.state_code,
													customer_address: selected.billing_address,
													customer_shipping_address: selected.shipping_address || selected.billing_address
												};
												recalculateTotals(updated);
											}
										}}
										style={{ height: '46px' }}
									>
										<option value="">-- Choose Customer --</option>
										{customers.map(c => (
											<option key={c.id} value={c.id}>{c.legal_name} ({c.gstin})</option>
										))}
									</select>
								</div>

								{/* Client Quick Details Display */}
								{formInvoice.customer_gstin ? (
									<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', width: '100%', background: 'var(--bg-input)', padding: '16px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
										<div style={{ gridColumn: 'span 2', display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border-color)', paddingBottom: '6px' }}>
											<strong style={{ fontSize: '0.95rem', color: 'var(--text-primary)' }}>{formInvoice.customer_name}</strong>
											<span className="client-badge">GST ACTIVE</span>
										</div>

										{/* Billing Address Block */}
										<div>
											<div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '4px' }}>
												<span style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Billing Address</span>
												<button
													type="button"
													onClick={() => setIsEditingBilling(!isEditingBilling)}
													style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '2px', color: 'var(--accent-color)', display: 'flex', alignItems: 'center' }}
													title="Edit Billing Address"
												>
													<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
														<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
														<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
													</svg>
												</button>
											</div>
											{isEditingBilling ? (
												<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
													<textarea
														className="b2b-input"
														style={{ fontSize: '0.85rem', minHeight: '60px', padding: '6px', width: '100%', boxSizing: 'border-box' }}
														value={formInvoice.customer_address}
														onChange={(e) => setFormInvoice({ ...formInvoice, customer_address: e.target.value })}
													/>
													<button
														type="button"
														onClick={() => setIsEditingBilling(false)}
														className="b2b-btn b2b-btn-primary"
														style={{ padding: '2px 8px', fontSize: '0.75rem', height: '22px', width: 'fit-content' }}
													>
														Save
													</button>
												</div>
											) : (
												<div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', lineHeight: 1.4, whiteSpace: 'pre-wrap' }}>
													{formInvoice.customer_address || <span style={{ fontStyle: 'italic', color: 'var(--text-tertiary)' }}>No billing address set</span>}
												</div>
											)}
										</div>

										{/* Shipping Address Block */}
										<div>
											<div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '4px' }}>
												<span style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Shipping Address</span>
												<button
													type="button"
													onClick={() => setIsEditingShipping(!isEditingShipping)}
													style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '2px', color: 'var(--accent-color)', display: 'flex', alignItems: 'center' }}
													title="Edit Shipping Address"
												>
													<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
														<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
														<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
													</svg>
												</button>
											</div>
											{isEditingShipping ? (
												<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
													<textarea
														className="b2b-input"
														style={{ fontSize: '0.85rem', minHeight: '60px', padding: '6px', width: '100%', boxSizing: 'border-box' }}
														value={formInvoice.customer_shipping_address || ''}
														onChange={(e) => setFormInvoice({ ...formInvoice, customer_shipping_address: e.target.value })}
													/>
													<button
														type="button"
														onClick={() => setIsEditingShipping(false)}
														className="b2b-btn b2b-btn-primary"
														style={{ padding: '2px 8px', fontSize: '0.75rem', height: '22px', width: 'fit-content' }}
													>
														Save
													</button>
												</div>
											) : (
												<div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', lineHeight: 1.4, whiteSpace: 'pre-wrap' }}>
													{formInvoice.customer_shipping_address || <span style={{ fontStyle: 'italic', color: 'var(--text-tertiary)' }}>No shipping address set</span>}
												</div>
											)}
										</div>

										{/* GSTIN / Supply State */}
										<div style={{ gridColumn: 'span 2', display: 'flex', gap: '16px', borderTop: '1px solid var(--border-color)', paddingTop: '8px', marginTop: '4px', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
											<div><strong>GSTIN:</strong> <span style={{ fontFamily: 'monospace', fontWeight: 600, color: 'var(--text-primary)' }}>{formInvoice.customer_gstin}</span></div>
											<div><strong>State:</strong> <span>{formInvoice.customer_state} ({formInvoice.customer_state_code})</span></div>
										</div>
									</div>
								) : (
									<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100px', border: '1px dashed var(--border-color)', borderRadius: '12px', background: 'var(--bg-hover)', color: 'var(--text-tertiary)', fontSize: '0.85rem', width: '100%' }}>
										<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ marginBottom: '8px', opacity: 0.6 }}>
											<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
										</svg>
										No client selected. Please choose a client to load details.
									</div>
								)}
							</div>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Billing Details</div>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', marginBottom: '20px' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Invoice Date*</label>
									<input
										type="date"
										className="b2b-input"
										value={formInvoice.invoice_date ? formInvoice.invoice_date.split('T')[0] : ''}
										onChange={(e) => {
											const newDate = e.target.value;
											let computedDueDate = formInvoice.due_date;
											if (formInvoice.terms) {
												const matchedTerm = paymentTerms.find(t => t.name === formInvoice.terms);
												if (matchedTerm) {
													const date = new Date(newDate);
													date.setDate(date.getDate() + matchedTerm.due_days);
													computedDueDate = date.toISOString().split('T')[0];
												}
											}
											// Refresh the next invoice number preview when the date changes
											if (viewMode === 'create') fetchNextInvoiceNumber(newDate);
											setFormInvoice({
												...formInvoice,
												invoice_date: newDate,
												due_date: computedDueDate
											});
										}}
									/>
								</div>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Due Date</label>
									<input
										type="date"
										className="b2b-input"
										value={formInvoice.due_date ? formInvoice.due_date.split('T')[0] : ''}
										onChange={(e) => setFormInvoice({ ...formInvoice, due_date: e.target.value })}
									/>
								</div>
							</div>

							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '20px', marginBottom: '20px' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Invoice Number</label>
									<div style={{
										display: 'flex',
										alignItems: 'center',
										gap: '10px',
										height: '46px',
										padding: '0 14px',
										borderRadius: '10px',
										border: '1.5px solid rgba(99,102,241,0.35)',
										background: 'linear-gradient(135deg, rgba(99,102,241,0.07), rgba(139,92,246,0.07))',
										boxShadow: '0 0 0 0 transparent',
										boxSizing: 'border-box',
									}}>
										<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="rgba(99,102,241,0.8)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0 }}>
											<rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
											<path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
										</svg>
										{viewMode === 'create' ? (
											<span style={{ fontFamily: 'monospace', fontWeight: 700, fontSize: '0.95rem', color: nextInvoiceNumber ? 'rgb(99,102,241)' : 'var(--text-tertiary)', letterSpacing: '0.5px' }}>
												{nextInvoiceNumber || 'Auto-generated on issue…'}
											</span>
										) : (
											<span style={{ fontFamily: 'monospace', fontWeight: 700, fontSize: '0.95rem', color: formInvoice.invoice_number ? 'rgb(16,185,129)' : 'var(--text-tertiary)', letterSpacing: '0.5px' }}>
												{formInvoice.invoice_number || 'DRAFT — assigned on issue'}
											</span>
										)}
										<span style={{ marginLeft: 'auto', fontSize: '0.7rem', color: 'var(--text-tertiary)', whiteSpace: 'nowrap' }}>read-only</span>
									</div>
								</div>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Salesperson</label>
									<input
										type="text"
										className="b2b-input"
										placeholder="e.g. John Doe"
										value={formInvoice.salesperson || ''}
										onChange={(e) => setFormInvoice({ ...formInvoice, salesperson: e.target.value })}
									/>
								</div>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Terms</label>
									<select
										className="b2b-input"
										value={formInvoice.terms || ''}
										onChange={(e) => {
											const val = e.target.value;
											if (val === '__CREATE_NEW__') {
												setShowPaymentTermModal(true);
											} else {
												handleTermsChange(val);
											}
										}}
										style={{ height: '46px' }}
									>
										<option value="">-- Select Terms --</option>
										{paymentTerms.map(t => (
											<option key={t.id || t.name} value={t.name}>{t.name}</option>
										))}
										<option value="__CREATE_NEW__" style={{ fontWeight: 'bold', color: 'var(--accent-color)' }}>+ New Payment Term</option>
									</select>
								</div>
							</div>

							<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
								<label className="form-label">Subject</label>
								<input
									type="text"
									className="b2b-input"
									placeholder="e.g. Supply of Raw materials / Fragrance ingredients"
									value={formInvoice.subject || ''}
									onChange={(e) => setFormInvoice({ ...formInvoice, subject: e.target.value })}
								/>
							</div>
						</div>

						{/* Item Table */}
						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Items Table</div>
							<div className="table-responsive" style={{ overflowX: 'auto', marginBottom: '20px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
								<table style={{ minWidth: '1220px', width: '100%', borderCollapse: 'collapse' }}>
									<thead>
										<tr className="items-table-header" style={{ borderBottom: '2px solid var(--border-color)', textAlign: 'left', background: 'var(--bg-hover)' }}>
											<th style={{ padding: '14px 16px', minWidth: '400px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>Item Details</th>
											<th style={{ padding: '14px 16px', minWidth: '140px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>HSN Code</th>
											<th style={{ padding: '14px 16px', minWidth: '120px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>GST (%)</th>
											<th style={{ padding: '14px 16px', minWidth: '110px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>Quantity</th>
											<th style={{ padding: '14px 16px', minWidth: '130px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>Rate (₹)</th>
											<th style={{ padding: '14px 16px', minWidth: '130px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>GST (₹)</th>
											<th style={{ padding: '14px 16px', minWidth: '140px', color: 'var(--text-secondary)', fontWeight: 600, textAlign: 'left' }}>Amount</th>
											<th style={{ padding: '14px 16px', minWidth: '50px', textAlign: 'right' }}></th>
										</tr>
									</thead>
									<tbody>
										{formInvoice.items.map((item, idx) => (
											<tr key={idx} className="items-table-row" style={{ borderBottom: '1px solid var(--border-color)' }}>
												<td style={{ padding: '12px 16px', minWidth: '400px' }}>
													<input
														type="text"
														list={`products-list-${idx}`}
														className="b2b-input"
														placeholder="Search warehouse products or type custom details..."
														value={item.item_details}
														onChange={(e) => {
															const val = e.target.value;
															const newItems = [...formInvoice.items];
															newItems[idx].item_details = val;

															// Check if matching a product option value or title
															const foundProduct = inventoryProducts.find(p =>
																p.title === val ||
																`${p.title} (${p.mi_sku})` === val
															);
															if (foundProduct) {
																newItems[idx].product_id = foundProduct.id;
																newItems[idx].item_details = foundProduct.title;
																newItems[idx].sku = foundProduct.mi_sku;
																newItems[idx].rate = foundProduct.price || 0;
																newItems[idx].hsn_code = foundProduct.hsn_code || '33029019';
															}
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														style={{ background: 'var(--surface-color)', height: '42px' }}
													/>
													<datalist id={`products-list-${idx}`}>
														{inventoryProducts.map(p => (
															<option key={p.id} value={`${p.title} (${p.mi_sku})`} />
														))}
													</datalist>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '140px' }}>
													<input
														type="text"
														className="b2b-input"
														placeholder="33029019"
														value={item.hsn_code || ''}
														onChange={(e) => {
															const newItems = [...formInvoice.items];
															newItems[idx].hsn_code = e.target.value;
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														maxLength={8}
														style={{ background: 'var(--surface-color)', height: '42px', fontFamily: 'monospace' }}
													/>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '120px' }}>
													<select
														className="b2b-input gst-select-input"
														value={item.gst_rate !== undefined ? item.gst_rate : 18}
														onChange={(e) => {
															const newItems = [...formInvoice.items];
															newItems[idx].gst_rate = Number(e.target.value);
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														style={{ background: 'var(--surface-color)', height: '42px' }}
													>
														<option value={0}>0%</option>
														<option value={5}>5%</option>
														<option value={12}>12%</option>
														<option value={18}>18%</option>
														<option value={28}>28%</option>
													</select>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '110px' }}>
													<input
														type="number"
														value={item.quantity}
														min="1"
														className="b2b-input"
														onChange={(e) => {
															const newItems = [...formInvoice.items];
															newItems[idx].quantity = Number(e.target.value);
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														style={{ background: 'var(--surface-color)', height: '42px' }}
													/>
													{item.product_id && (
														<div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '4px', textAlign: 'center' }}>
															Stock: <span style={{ fontWeight: 600, color: (getProductStock(item.product_id) || 0) <= 0 ? 'var(--status-danger)' : 'var(--text-secondary)' }}>{getProductStock(item.product_id) ?? 0}</span>
														</div>
													)}
												</td>
												<td style={{ padding: '12px 16px', minWidth: '130px' }}>
													<input
														type="number"
														value={item.rate}
														className="b2b-input"
														onChange={(e) => {
															const newItems = [...formInvoice.items];
															newItems[idx].rate = Number(e.target.value);
															recalculateTotals({ ...formInvoice, items: newItems });
														}}
														style={{ background: 'var(--surface-color)', height: '42px' }}
													/>
												</td>
												<td style={{ padding: '12px 16px', minWidth: '130px', fontWeight: '500', color: 'var(--text-secondary)', whiteSpace: 'nowrap', fontSize: '0.9rem' }}>
													₹{((item.quantity * item.rate * (item.gst_rate !== undefined ? item.gst_rate : 18)) / 100).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
												</td>
												<td style={{ padding: '12px 16px', minWidth: '140px', fontWeight: '700', color: 'var(--text-primary)', whiteSpace: 'nowrap', fontSize: '0.95rem' }}>₹{(item.quantity * item.rate).toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
												<td style={{ padding: '12px 16px', minWidth: '50px', textAlign: 'right' }}>
													{formInvoice.items.length > 1 && (
														<button
															onClick={() => {
																const newItems = formInvoice.items.filter((_, i) => i !== idx);
																recalculateTotals({ ...formInvoice, items: newItems });
															}}
															className="delete-row-btn"
															title="Delete Row"
														>
															<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
																<polyline points="3 6 5 6 21 6" />
																<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
																<line x1="10" y1="11" x2="10" y2="17" />
																<line x1="14" y1="11" x2="14" y2="17" />
															</svg>
														</button>
													)}
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>

							<button
								onClick={() => {
									const newItems = [...formInvoice.items, { item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019', gst_rate: 18 }];
									setFormInvoice({ ...formInvoice, items: newItems });
								}}
								className="add-row-btn"
							>
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
								</svg>
								Add Row
							</button>
						</div>

						{/* Calculations section */}
						<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', marginTop: '24px' }}>
							{/* Notes & Extra charges selection */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Customer Notes</label>
									<textarea
										placeholder="Thanks for your business. Let us know if you need any changes."
										value={formInvoice.customer_notes || ''}
										onChange={(e) => setFormInvoice({ ...formInvoice, customer_notes: e.target.value })}
										className="b2b-input"
										style={{ minHeight: '110px', fontFamily: 'inherit' }}
									/>
								</div>

								<div style={{ background: 'var(--bg-hover)', padding: '20px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
									<div style={{ fontWeight: 700, fontSize: '0.85rem', color: 'var(--text-secondary)', textTransform: 'uppercase', marginBottom: '14px', letterSpacing: '0.05em' }}>Tax Adjustments</div>
									<div style={{ display: 'flex', gap: '20px', alignItems: 'center', flexWrap: 'wrap' }}>

										<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
											<label className="form-label" style={{ fontSize: '0.7rem' }}>TDS/TCS Type</label>
											<select
												value={formInvoice.tds_tcs_type}
												onChange={(e) => {
													const type = e.target.value;
													const rate = type === 'NONE' ? 0 : formInvoice.tds_tcs_rate;
													recalculateTotals({ ...formInvoice, tds_tcs_type: type, tds_tcs_rate: rate });
												}}
												className="b2b-input"
												style={{ width: '130px', height: '40px' }}
											>
												<option value="NONE">None</option>
												<option value="TDS">TDS</option>
												<option value="TCS">TCS</option>
											</select>
										</div>
										{formInvoice.tds_tcs_type !== 'NONE' && (
											<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
												<label className="form-label" style={{ fontSize: '0.7rem' }}>Rate (%)</label>
												<input
													type="number"
													value={formInvoice.tds_tcs_rate}
													onChange={(e) => recalculateTotals({ ...formInvoice, tds_tcs_rate: Number(e.target.value) })}
													className="b2b-input"
													style={{ width: '100px', height: '40px' }}
												/>
											</div>
										)}
									</div>
								</div>
							</div>

							{/* Pricing Breakdown */}
							<div className="summary-panel">
								<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
									<span>Sub Total:</span>
									<strong style={{ color: 'var(--text-primary)' }}>₹{formInvoice.subtotal_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</strong>
								</div>

								<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
									<span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>Discount (%):</span>
									<input
										type="number"
										value={formInvoice.discount_percent}
										min="0"
										max="100"
										className="b2b-input"
										onChange={(e) => recalculateTotals({ ...formInvoice, discount_percent: Number(e.target.value) })}
										style={{ width: '70px', padding: '6px 10px !important', textAlign: 'right', background: 'var(--surface-color)', height: '32px' }}
									/>
								</div>

								{formInvoice.discount_amount > 0 && (
									<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
										<span>Discount Amount:</span>
										<span style={{ color: 'var(--status-danger)', fontWeight: 600 }}>- ₹{formInvoice.discount_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								{formInvoice.cgst_amount > 0 && (
									<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
										<span>CGST ({formInvoice.cgst_rate}%):</span>
										<span>₹{formInvoice.cgst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}
								{formInvoice.sgst_amount > 0 && (
									<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
										<span>SGST ({formInvoice.sgst_rate}%):</span>
										<span>₹{formInvoice.sgst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}
								{formInvoice.igst_amount > 0 && (
									<div className="summary-row" style={{ color: 'var(--text-secondary)' }}>
										<span>IGST ({formInvoice.igst_rate}%):</span>
										<span>₹{formInvoice.igst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								<div className="summary-row" style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
									<span>Transportation:</span>
									<input
										type="number"
										value={formInvoice.transportation_charge}
										className="b2b-input"
										onChange={(e) => recalculateTotals({ ...formInvoice, transportation_charge: Number(e.target.value) })}
										style={{ width: '110px', padding: '6px 10px !important', textAlign: 'right', background: 'var(--surface-color)', height: '32px' }}
									/>
								</div>

								{formInvoice.tds_tcs_type !== 'NONE' && (
									<div className="summary-row" style={{ color: 'var(--status-warning)', fontWeight: 600 }}>
										<span>{formInvoice.tds_tcs_type} ({formInvoice.tds_tcs_rate}%):</span>
										<span>{formInvoice.tds_tcs_type === 'TDS' ? '-' : '+'} ₹{formInvoice.tds_tcs_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', margin: '16px 0' }} />

								<div className="summary-total-box">
									<span style={{ color: 'var(--text-primary)', fontWeight: 700, fontSize: '1rem' }}>Total Amount:</span>
									<span style={{ color: 'var(--accent-color)', fontSize: '1.4rem', fontWeight: 800 }}>₹{formInvoice.total_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
								</div>
							</div>
						</div>

						{/* Form Buttons */}
						<div style={{ display: 'flex', justifyContent: 'flex-end', gap: '16px', marginTop: '40px', borderTop: '1px solid var(--border-color)', paddingTop: '28px' }}>
							<button
								onClick={() => setViewMode('list')}
								className="b2b-btn b2b-btn-secondary"
								style={{ minWidth: '110px', height: '46px' }}
							>
								Cancel
							</button>
							<button
								onClick={() => handleSaveInvoice(true)}
								className="b2b-btn b2b-btn-secondary"
								style={{ minWidth: '140px', height: '46px' }}
							>
								Save as Draft
							</button>
							<button
								onClick={() => handleSaveInvoice(false)}
								className="b2b-btn b2b-btn-primary"
								style={{ minWidth: '180px', height: '46px', gap: '8px' }}
							>
								<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<line x1="22" y1="2" x2="11" y2="13" /><polygon points="22 2 15 22 11 13 2 9 22 2" />
								</svg>
								Save & Issue Bill
							</button>
						</div>
					</div>
				)}

				{/* PROFORMA CREATOR & EDITOR VIEW */}
				{(viewMode === 'create-pf' || viewMode === 'edit-pf') && (
					<div className="b2b-form-area no-print" style={{ color: 'var(--text-primary)' }}>
						<div className="form-header-container">
							<div>
								<h3 className="form-header-title">
									<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--accent-color)' }}>
										<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
										<polyline points="14 2 14 8 20 8" />
										<line x1="16" y1="13" x2="8" y2="13" />
										<line x1="16" y1="17" x2="8" y2="17" />
										<polyline points="10 9 9 9 8 9" />
									</svg>
									{viewMode === 'create-pf' ? 'Create Proforma Invoice' : `Edit Proforma Invoice (Rev ${formProforma.revision_number})`}
								</h3>
								<p className="form-header-subtitle">Fill in the commercial and pricing estimates for the client approval stage.</p>
							</div>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary" style={{ height: '40px' }}>
								&larr; Back to List
							</button>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Client Information</div>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', alignItems: 'start' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Select B2B Customer*</label>
									<select
										className="b2b-input"
										value={formProforma.customer_id || ''}
										onChange={(e) => {
											const cId = Number(e.target.value);
											const selected = customers.find(c => c.id === cId);
											if (selected) {
												const updated = {
													...formProforma,
													customer_id: selected.id,
													customer_gstin: selected.gstin,
													customer_name: selected.legal_name,
													customer_state: selected.state,
													customer_state_code: selected.state_code,
													customer_address: selected.billing_address,
													customer_shipping_address: selected.shipping_address || selected.billing_address
												};
												recalculateProformaTotals(updated);
											}
										}}
										style={{ height: '46px' }}
									>
										<option value="">-- Choose Customer --</option>
										{customers.map(c => (
											<option key={c.id} value={c.id}>{c.legal_name} ({c.gstin})</option>
										))}
									</select>
								</div>

								{formProforma.customer_gstin ? (
									<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', width: '100%', background: 'var(--bg-input)', padding: '16px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
										<div style={{ gridColumn: 'span 2', display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border-color)', paddingBottom: '6px' }}>
											<strong style={{ fontSize: '0.95rem', color: 'var(--text-primary)' }}>{formProforma.customer_name}</strong>
											<span className="client-badge">PROSPECTIVE</span>
										</div>

										{/* Billing Address Block */}
										<div>
											<div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '4px' }}>
												<span style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Billing Address</span>
												<button
													type="button"
													onClick={() => setIsEditingBilling(!isEditingBilling)}
													style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '2px', color: 'var(--accent-color)', display: 'flex', alignItems: 'center' }}
													title="Edit Billing Address"
												>
													<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
														<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
														<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
													</svg>
												</button>
											</div>
											{isEditingBilling ? (
												<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
													<textarea
														className="b2b-input"
														style={{ fontSize: '0.85rem', minHeight: '60px', padding: '6px', width: '100%', boxSizing: 'border-box' }}
														value={formProforma.customer_address}
														onChange={(e) => setFormProforma({ ...formProforma, customer_address: e.target.value })}
													/>
													<button
														type="button"
														onClick={() => setIsEditingBilling(false)}
														className="b2b-btn b2b-btn-primary"
														style={{ padding: '2px 8px', fontSize: '0.75rem', height: '22px', width: 'fit-content' }}
													>
														Save
													</button>
												</div>
											) : (
												<div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', lineHeight: 1.4, whiteSpace: 'pre-wrap' }}>
													{formProforma.customer_address || <span style={{ fontStyle: 'italic', color: 'var(--text-tertiary)' }}>No billing address set</span>}
												</div>
											)}
										</div>

										{/* Shipping Address Block */}
										<div>
											<div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '4px' }}>
												<span style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Shipping Address</span>
												<button
													type="button"
													onClick={() => setIsEditingShipping(!isEditingShipping)}
													style={{ background: 'none', border: 'none', cursor: 'pointer', padding: '2px', color: 'var(--accent-color)', display: 'flex', alignItems: 'center' }}
													title="Edit Shipping Address"
												>
													<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
														<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
														<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
													</svg>
												</button>
											</div>
											{isEditingShipping ? (
												<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
													<textarea
														className="b2b-input"
														style={{ fontSize: '0.85rem', minHeight: '60px', padding: '6px', width: '100%', boxSizing: 'border-box' }}
														value={formProforma.customer_shipping_address || ''}
														onChange={(e) => setFormProforma({ ...formProforma, customer_shipping_address: e.target.value })}
													/>
													<button
														type="button"
														onClick={() => setIsEditingShipping(false)}
														className="b2b-btn b2b-btn-primary"
														style={{ padding: '2px 8px', fontSize: '0.75rem', height: '22px', width: 'fit-content' }}
													>
														Save
													</button>
												</div>
											) : (
												<div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', lineHeight: 1.4, whiteSpace: 'pre-wrap' }}>
													{formProforma.customer_shipping_address || <span style={{ fontStyle: 'italic', color: 'var(--text-tertiary)' }}>No shipping address set</span>}
												</div>
											)}
										</div>

										{/* GSTIN / Supply State */}
										<div style={{ gridColumn: 'span 2', display: 'flex', gap: '16px', borderTop: '1px solid var(--border-color)', paddingTop: '8px', marginTop: '4px', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
											<div><strong>GSTIN:</strong> <span style={{ fontFamily: 'monospace', fontWeight: 600, color: 'var(--text-primary)' }}>{formProforma.customer_gstin}</span></div>
											<div><strong>State:</strong> <span>{formProforma.customer_state} ({formProforma.customer_state_code})</span></div>
										</div>
									</div>
								) : (
									<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100px', border: '1px dashed var(--border-color)', borderRadius: '12px', background: 'var(--bg-hover)', color: 'var(--text-tertiary)', fontSize: '0.85rem', width: '100%' }}>
										No client selected. Choose a B2B customer to proceed.
									</div>
								)}
							</div>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Validity & Date Details</div>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Proforma Date*</label>
									<input
										type="date"
										className="b2b-input"
										value={formProforma.note_date ? formProforma.note_date.split('T')[0] : ''}
										onChange={(e) => {
											const newDate = e.target.value;
											setFormProforma({
												...formProforma,
												note_date: newDate
											});
											if (viewMode === 'create-pf' && !formProforma.parent_proforma_id) {
												fetchNextProformaNumber(newDate);
											}
										}}
										style={{ height: '46px' }}
									/>
								</div>

								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Valid Until</label>
									<input
										type="date"
										className="b2b-input"
										value={formProforma.valid_until ? formProforma.valid_until.split('T')[0] : ''}
										onChange={(e) => {
											setFormProforma({
												...formProforma,
												valid_until: e.target.value
											});
										}}
										style={{ height: '46px' }}
									/>
								</div>
							</div>
						</div>

						<div className="b2b-form-section">
							<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
								<div className="b2b-form-section-title" style={{ margin: 0 }}>Item Estimations</div>
								<button
									onClick={() => {
										const items = [...formProforma.items, { item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }];
										recalculateProformaTotals({ ...formProforma, items });
									}}
									className="b2b-btn b2b-btn-secondary"
									style={{ padding: '6px 12px', fontSize: '0.85rem' }}
								>
									+ Add Line Item
								</button>
							</div>

							<div className="table-responsive" style={{ overflowX: 'auto', border: '1px solid var(--border-color)', borderRadius: '12px' }}>
								<table className="b2b-table">
									<thead>
										<tr style={{ background: 'var(--bg-hover)' }}>
											<th>Product Details / Description*</th>
											<th style={{ width: '130px' }}>HSN Code</th>
											<th style={{ width: '100px', textAlign: 'right' }}>Qty</th>
											<th style={{ width: '140px', textAlign: 'right' }}>Rate (₹)</th>
											<th style={{ width: '100px', textAlign: 'center' }}>GST %</th>
											<th style={{ width: '140px', textAlign: 'right' }}>Amount (₹)</th>
											<th style={{ width: '60px', textAlign: 'center' }}></th>
										</tr>
									</thead>
									<tbody>
										{formProforma.items.map((item: any, idx: number) => (
											<tr key={idx}>
												<td>
													<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
														<select
															className="b2b-input"
															value={item.product_id || ''}
															onChange={(e) => {
																const pId = Number(e.target.value);
																const prod = inventoryProducts.find(p => p.id === pId);
																if (prod) {
																	const newItems = [...formProforma.items];
																	newItems[idx] = {
																		...newItems[idx],
																		product_id: prod.id,
																		item_details: prod.title,
																		sku: prod.mi_sku,
																		hsn_code: prod.hsn_code || '33029019',
																		rate: prod.price || 0
																	};
																	recalculateProformaTotals({ ...formProforma, items: newItems });
																}
															}}
														>
															<option value="">-- Autoload from Inventory --</option>
															{inventoryProducts.map(p => (
																<option key={p.id} value={p.id}>{p.title} ({p.mi_sku})</option>
															))}
														</select>
														<input
															type="text"
															className="b2b-input"
															placeholder="Custom Description"
															value={item.item_details}
															onChange={(e) => {
																const newItems = [...formProforma.items];
																newItems[idx].item_details = e.target.value;
																setFormProforma({ ...formProforma, items: newItems });
															}}
														/>
													</div>
												</td>
												<td>
													<input
														type="text"
														className="b2b-input"
														placeholder="HSN"
														value={item.hsn_code || ''}
														onChange={(e) => {
															const newItems = [...formProforma.items];
															newItems[idx].hsn_code = e.target.value;
															setFormProforma({ ...formProforma, items: newItems });
														}}
													/>
												</td>
												<td>
													<input
														type="number"
														className="b2b-input"
														style={{ textAlign: 'right' }}
														value={item.quantity}
														onChange={(e) => {
															const newItems = [...formProforma.items];
															newItems[idx].quantity = Number(e.target.value);
															recalculateProformaTotals({ ...formProforma, items: newItems });
														}}
													/>
													{item.product_id && (
														<div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '4px', textAlign: 'center' }}>
															Stock: <span style={{ fontWeight: 600, color: (getProductStock(item.product_id) || 0) <= 0 ? 'var(--status-danger)' : 'var(--text-secondary)' }}>{getProductStock(item.product_id) ?? 0}</span>
														</div>
													)}
												</td>
												<td>
													<input
														type="number"
														className="b2b-input"
														style={{ textAlign: 'right' }}
														value={item.rate}
														onChange={(e) => {
															const newItems = [...formProforma.items];
															newItems[idx].rate = Number(e.target.value);
															recalculateProformaTotals({ ...formProforma, items: newItems });
														}}
													/>
												</td>
												<td>
													<select
														className="b2b-input"
														style={{ textAlign: 'center' }}
														value={item.gst_rate !== undefined ? item.gst_rate : 18}
														onChange={(e) => {
															const newItems = [...formProforma.items];
															newItems[idx].gst_rate = Number(e.target.value);
															recalculateProformaTotals({ ...formProforma, items: newItems });
														}}
													>
														<option value="18">18%</option>
														<option value="12">12%</option>
														<option value="28">28%</option>
														<option value="5">5%</option>
														<option value="0">0%</option>
													</select>
												</td>
												<td style={{ textAlign: 'right', fontWeight: '500', paddingRight: '12px' }}>
													₹{(item.amount || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })}
												</td>
												<td style={{ textAlign: 'center' }}>
													<button
														type="button"
														aria-label="Remove item"
														disabled={formProforma.items.length <= 1}
														onClick={() => {
															const newItems = formProforma.items.filter((_: any, i: number) => i !== idx);
															recalculateProformaTotals({ ...formProforma, items: newItems });
														}}
														className="clear-search-btn"
														style={{ display: 'inline-block', position: 'static', opacity: formProforma.items.length <= 1 ? 0.3 : 1 }}
													>
														✕
													</button>
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
						</div>

						<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', marginTop: '24px' }}>
							<div>
								<div className="b2b-form-section">
									<div className="b2b-form-section-title">Advance Deposit Request</div>
									<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
										<label className="form-label">Advance Payment Recorded (₹)</label>
										<input
											type="number"
											className="b2b-input"
											value={formProforma.advance_paid || 0}
											onChange={(e) => {
												setFormProforma({
													...formProforma,
													advance_paid: Number(e.target.value)
												});
											}}
											style={{ height: '46px' }}
											placeholder="Enter advance deposit amount received if any"
										/>
									</div>
								</div>
							</div>

							<div style={{ background: 'var(--bg-hover)', border: '1px solid var(--border-color)', borderRadius: '16px', padding: '24px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
								<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px dashed var(--border-color)', paddingBottom: '10px' }}>
									<span>Subtotal:</span>
									<span style={{ fontWeight: '500' }}>₹{formProforma.subtotal_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
								</div>

								<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px dashed var(--border-color)', paddingBottom: '10px' }}>
									<div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
										<span>Discount Percent:</span>
										<input
											type="number"
											className="b2b-input"
											style={{ width: '60px', height: '32px', textAlign: 'center', padding: '2px' }}
											value={formProforma.discount_percent || 0}
											onChange={(e) => recalculateProformaTotals({ ...formProforma, discount_percent: Number(e.target.value) })}
										/>
										<span>%</span>
									</div>
									<span style={{ color: 'var(--danger-color)' }}>-₹{formProforma.discount_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
								</div>

								{formProforma.cgst_amount > 0 && (
									<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px dashed var(--border-color)', paddingBottom: '10px', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
										<span>CGST ({formProforma.cgst_rate}%):</span>
										<span>₹{formProforma.cgst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								{formProforma.sgst_amount > 0 && (
									<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px dashed var(--border-color)', paddingBottom: '10px', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
										<span>SGST ({formProforma.sgst_rate}%):</span>
										<span>₹{formProforma.sgst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								{formProforma.igst_amount > 0 && (
									<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px dashed var(--border-color)', paddingBottom: '10px', fontSize: '0.9rem', color: 'var(--text-secondary)' }}>
										<span>IGST ({formProforma.igst_rate}%):</span>
										<span>₹{formProforma.igst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
									</div>
								)}

								<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '1.25rem', fontWeight: '700', color: 'var(--accent-color)', paddingTop: '10px' }}>
									<span>Estimated Total:</span>
									<span>₹{formProforma.total_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
								</div>
							</div>
						</div>

						<div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '30px', borderTop: '1px solid var(--border-color)', paddingTop: '20px' }}>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary" style={{ width: '120px', height: '46px' }}>
								Cancel
							</button>
							<button
								onClick={() => handleSaveProforma(true)}
								className="b2b-btn b2b-btn-secondary"
								style={{ minWidth: '150px', height: '46px', background: 'var(--bg-hover)' }}
							>
								Save as Draft
							</button>
							<button
								onClick={() => handleSaveProforma(false)}
								className="b2b-btn b2b-btn-primary"
								style={{ minWidth: '180px', height: '46px', gap: '8px' }}
							>
								Save & Issue Proforma
							</button>
						</div>
					</div>
				)}

				{/* PRINT / PREVIEW LAYOUT VIEW */}
				{viewMode === 'preview' && selectedInvoice && (
					<div>
						<div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '24px', borderBottom: '1px solid var(--border-color)', paddingBottom: '12px', alignItems: 'center' }}>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary">&larr; Back to List</button>
							<div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
								{selectedInvoice.status === 'DRAFT' && userRole === 'admin' && (
									<button
										onClick={() => {
											setFormInvoice({ ...selectedInvoice });
											setViewMode('edit');
										}}
										className="b2b-btn b2b-btn-secondary"
									>
										Edit Invoice
									</button>
								)}
								{selectedInvoice.status === 'ISSUED' && userRole === 'admin' && (
									<>
										{selectedInvoice.inventory_deducted ? (
											<div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
												<span style={{ display: 'inline-flex', alignItems: 'center', background: 'var(--status-active-bg)', color: 'var(--status-active)', padding: '6px 12px', borderRadius: '8px', fontSize: '13px', fontWeight: 600 }}>
													✓ Stock Deducted
												</span>
												<button
													onClick={() => handleRevertInventory(selectedInvoice.id!)}
													className="b2b-btn"
													style={{ background: 'var(--status-danger-bg)', color: 'var(--status-danger)', border: '1px solid var(--status-danger)' }}
												>
													Revert Stock
												</button>
											</div>
										) : (
											<button
												onClick={() => handleDeductInventory(selectedInvoice.id!)}
												className="b2b-btn"
												style={{ background: '#eab308', color: '#000', border: 'none', fontWeight: 600 }}
											>
												Deduct Inventory
											</button>
										)}
									</>
								)}
								<button onClick={triggerPrint} className="b2b-btn b2b-btn-primary">Print / Download PDF</button>
							</div>
						</div>

						{/* INVOICE PAGE DESIGN */}
						<div className="print-invoice-area" style={{ background: 'var(--surface-color)', padding: '40px', borderRadius: '16px', border: '1px solid var(--border-color)', color: 'var(--text-primary)', boxShadow: 'var(--shadow-sm)' }}>
							<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '36px' }}>
								<div>
									<h1 style={{ margin: '0 0 8px 0', textTransform: 'uppercase', color: 'var(--accent-color)', fontWeight: 800, fontSize: '2rem' }}>TAX INVOICE</h1>
									<div style={{ fontSize: '14px', color: 'var(--text-secondary)' }}><strong>FY:</strong> {selectedInvoice.financial_year || 'N/A'}</div>
								</div>
								<div style={{ textAlign: 'right' }}>
									<h2 style={{ margin: '0 0 6px 0', fontWeight: 800 }}>{selectedInvoice.seller_name}</h2>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedInvoice.seller_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', maxWidth: '280px', lineHeight: 1.4 }}>{selectedInvoice.seller_address}</div>
								</div>
							</div>

							<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', marginBottom: '24px' }} />

							<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1.2fr 1.1fr', gap: '30px', marginBottom: '36px' }}>
								<div>
									<h4 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontSize: '11px', letterSpacing: '0.05em' }}>Bill To</h4>
									<h3 style={{ margin: '0 0 6px 0', fontWeight: 700, fontSize: '15px' }}>{selectedInvoice.customer_name}</h3>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedInvoice.customer_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', lineHeight: 1.4 }}>{selectedInvoice.customer_address}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '4px' }}><strong>State:</strong> {selectedInvoice.customer_state} ({selectedInvoice.customer_state_code})</div>
								</div>
								<div>
									<h4 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontSize: '11px', letterSpacing: '0.05em' }}>Ship To</h4>
									<h3 style={{ margin: '0 0 6px 0', fontWeight: 700, fontSize: '15px' }}>{selectedInvoice.customer_name}</h3>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedInvoice.customer_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', lineHeight: 1.4 }}>{selectedInvoice.customer_shipping_address || selectedInvoice.customer_address}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '4px' }}><strong>State:</strong> {selectedInvoice.customer_state} ({selectedInvoice.customer_state_code})</div>
								</div>
								<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1.5fr', gap: '10px 16px', fontSize: '13px', color: 'var(--text-secondary)' }}>
									<strong>Invoice Number:</strong>
									<span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{selectedInvoice.invoice_number || 'DRAFT'}</span>

									<strong>Invoice Date:</strong>
									<span style={{ color: 'var(--text-primary)' }}>{selectedInvoice.invoice_date ? selectedInvoice.invoice_date.split('T')[0] : ''}</span>

									{selectedInvoice.due_date && (
										<>
											<strong>Due Date:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedInvoice.due_date.split('T')[0]}</span>
										</>
									)}

									{selectedInvoice.order_number && (
										<>
											<strong>Order Number:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedInvoice.order_number}</span>
										</>
									)}

									{selectedInvoice.salesperson && (
										<>
											<strong>Salesperson:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedInvoice.salesperson}</span>
										</>
									)}
								</div>
							</div>

							{selectedInvoice.subject && (
								<div style={{ background: 'var(--bg-input)', padding: '12px 16px', borderRadius: '8px', marginBottom: '32px', borderLeft: '4px solid var(--accent-color)', color: 'var(--text-primary)' }}>
									<strong>Subject:</strong> {selectedInvoice.subject}
								</div>
							)}

							<table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '36px' }}>
								<thead>
									<tr style={{ borderBottom: '2px solid var(--border-color)', textAlign: 'left', fontSize: '13px', color: 'var(--text-secondary)' }}>
										<th style={{ padding: '12px 8px', fontWeight: 600 }}>Item Details</th>
										<th style={{ padding: '12px 8px', width: '100px', textAlign: 'right', fontWeight: 600 }}>Qty</th>
										<th style={{ padding: '12px 8px', width: '120px', textAlign: 'right', fontWeight: 600 }}>Rate (₹)</th>
										<th style={{ padding: '12px 8px', width: '120px', textAlign: 'right', fontWeight: 600 }}>Amount (₹)</th>
									</tr>
								</thead>
								<tbody>
									{selectedInvoice.items.map((item, idx) => (
										<tr key={idx} style={{ borderBottom: '1px solid var(--border-color)', fontSize: '14px', color: 'var(--text-primary)' }}>
											<td style={{ padding: '12px 8px' }}>
												<strong>{item.item_details}</strong>
												{item.hsn_code && <span style={{ display: 'block', fontSize: '11px', color: 'var(--text-tertiary)', marginTop: '2px' }}>HSN: {item.hsn_code}</span>}
											</td>
											<td style={{ padding: '12px 8px', textAlign: 'right' }}>{item.quantity}</td>
											<td style={{ padding: '12px 8px', textAlign: 'right' }}>₹{item.rate.toFixed(2)}</td>
											<td style={{ padding: '12px 8px', textAlign: 'right', fontWeight: 'bold' }}>₹{item.amount.toFixed(2)}</td>
										</tr>
									))}
								</tbody>
							</table>

							{/* Calculations breakdowns */}
							<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px' }}>
								<div>
									{selectedInvoice.customer_notes && (
										<div>
											<h5 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', fontSize: '11px', textTransform: 'uppercase' }}>Customer Notes</h5>
											<div style={{ color: 'var(--text-secondary)', fontSize: '13px', lineHeight: 1.4, whiteSpace: 'pre-wrap' }}>{selectedInvoice.customer_notes}</div>
										</div>
									)}

									{/* Bank Details & UPI QR Code Section */}
									{(appConfigs['bank_name'] || appConfigs['bank_account_no'] || appConfigs['bank_ifsc'] || appConfigs['upi_id']) && (
										<div style={{ 
											marginTop: '24px', 
											padding: '20px', 
											borderRadius: '14px', 
											border: '1px solid rgba(14, 165, 233, 0.12)', 
											background: 'linear-gradient(135deg, rgba(14, 165, 233, 0.02) 0%, rgba(14, 165, 233, 0.05) 100%)',
											boxShadow: '0 4px 12px -2px rgba(14, 165, 233, 0.04)'
										}}>
											<div style={{ 
												display: 'flex', 
												justifyContent: 'space-between', 
												alignItems: 'center',
												borderBottom: '1px solid rgba(14, 165, 233, 0.1)', 
												paddingBottom: '8px', 
												marginBottom: '12px' 
											}}>
												<h5 style={{ margin: 0, fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em', color: 'var(--accent-color)', fontWeight: 700 }}>Payment Details</h5>
												{appConfigs['upi_id'] && (
													<span style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '10px', color: 'var(--success-color)', fontWeight: 600 }}>
														<span style={{ width: '6px', height: '6px', borderRadius: '50%', background: 'var(--success-color)', display: 'inline-block' }}></span>
														UPI Active
													</span>
												)}
											</div>
											<div style={{ display: 'flex', gap: '20px', alignItems: 'center' }}>
												<div style={{ flex: 1, fontSize: '13px', color: 'var(--text-secondary)' }}>
													{appConfigs['bank_name'] && (
														<div style={{ display: 'flex', marginBottom: '6px' }}>
															<span style={{ width: '90px', color: 'var(--text-tertiary)', fontWeight: 500 }}>Bank:</span>
															<span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{appConfigs['bank_name']}</span>
														</div>
													)}
													{appConfigs['bank_account_no'] && (
														<div style={{ display: 'flex', marginBottom: '6px' }}>
															<span style={{ width: '90px', color: 'var(--text-tertiary)', fontWeight: 500 }}>Account No:</span>
															<span style={{ fontFamily: 'monospace', fontWeight: 600, color: 'var(--text-primary)' }}>{appConfigs['bank_account_no']}</span>
														</div>
													)}
													{appConfigs['bank_ifsc'] && (
														<div style={{ display: 'flex', marginBottom: '6px' }}>
															<span style={{ width: '90px', color: 'var(--text-tertiary)', fontWeight: 500 }}>IFSC Code:</span>
															<span style={{ fontFamily: 'monospace', fontWeight: 600, color: 'var(--text-primary)' }}>{appConfigs['bank_ifsc']}</span>
														</div>
													)}
													{appConfigs['upi_id'] && (
														<div style={{ display: 'flex' }}>
															<span style={{ width: '90px', color: 'var(--text-tertiary)', fontWeight: 500 }}>UPI ID:</span>
															<span style={{ fontWeight: 600, color: 'var(--accent-color)' }}>{appConfigs['upi_id']}</span>
														</div>
													)}
												</div>
												{appConfigs['upi_id'] && selectedInvoice.balance_amount > 0 && (
													<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '6px' }}>
														<div style={{ background: 'white', padding: '6px', borderRadius: '8px', border: '1px solid rgba(14, 165, 233, 0.15)', boxShadow: '0 4px 10px rgba(0,0,0,0.04)' }}>
															<img
																src={`https://api.qrserver.com/v1/create-qr-code/?size=110x110&data=${encodeURIComponent(
																	`upi://pay?pa=${appConfigs['upi_id']}&pn=${encodeURIComponent(selectedInvoice.seller_name || '')}&am=${selectedInvoice.balance_amount.toFixed(2)}&cu=INR`
																)}`}
																alt="Payment QR"
																style={{ width: '100px', height: '100px', display: 'block' }}
															/>
														</div>
														<span style={{ fontSize: '10px', color: 'var(--text-tertiary)', fontWeight: 600 }}>Scan to Pay</span>
													</div>
												)}
											</div>
										</div>
									)}
								</div>
								<div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
									<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
										<span>Sub Total:</span>
										<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.subtotal_price.toFixed(2)}</span>
									</div>

									{selectedInvoice.discount_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>Discount ({selectedInvoice.discount_percent}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>- ₹{selectedInvoice.discount_amount.toFixed(2)}</span>
										</div>
									)}

									{selectedInvoice.cgst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>CGST ({selectedInvoice.cgst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.cgst_amount.toFixed(2)}</span>
										</div>
									)}
									{selectedInvoice.sgst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>SGST ({selectedInvoice.sgst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.sgst_amount.toFixed(2)}</span>
										</div>
									)}
									{selectedInvoice.igst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>IGST ({selectedInvoice.igst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.igst_amount.toFixed(2)}</span>
										</div>
									)}

									{selectedInvoice.transportation_charge > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>Transportation Charges:</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedInvoice.transportation_charge.toFixed(2)}</span>
										</div>
									)}

									{selectedInvoice.tds_tcs_type !== 'NONE' && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px', color: 'var(--status-warning)', fontWeight: 600 }}>
											<span>{selectedInvoice.tds_tcs_type} ({selectedInvoice.tds_tcs_rate}%):</span>
											<span>{selectedInvoice.tds_tcs_type === 'TDS' ? '-' : '+'} ₹{selectedInvoice.tds_tcs_amount.toFixed(2)}</span>
										</div>
									)}

									<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', margin: '12px 0' }} />

									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '18px', fontWeight: 800, color: 'var(--text-primary)' }}>
										<span>Grand Total (₹):</span>
										<span>₹{selectedInvoice.total_price.toFixed(2)}</span>
									</div>

									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '13px', marginTop: '8px' }}>
										<span>Paid Amount:</span>
										<span>₹{selectedInvoice.paid_amount.toFixed(2)}</span>
									</div>
									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '13px', fontWeight: 700, marginTop: '4px' }}>
										<span>Balance Due:</span>
										<span style={{ color: selectedInvoice.balance_amount > 0 ? 'var(--status-warning)' : 'var(--status-active)' }}>₹{selectedInvoice.balance_amount.toFixed(2)}</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				)}

				{/* PROFORMA PREVIEW & PRINT LAYOUT VIEW */}
				{viewMode === 'preview-pf' && selectedProforma && (() => {
					const rootId = selectedProforma.parent_proforma_id || selectedProforma.id;
					const historyStream = proformas
						.filter(p => p.id === rootId || p.parent_proforma_id === rootId)
						.sort((a, b) => a.revision_number - b.revision_number);

					return (
						<div>
							<div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '24px', borderBottom: '1px solid var(--border-color)', paddingBottom: '12px', flexWrap: 'wrap', gap: '12px', alignItems: 'center' }}>
								<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary">&larr; Back to List</button>
								<div style={{ display: 'flex', gap: '8px' }}>
									{selectedProforma.status === 'DRAFT' && userRole === 'admin' && (
										<button
											onClick={() => {
												setFormProforma({ ...selectedProforma });
												setViewMode('edit-pf');
											}}
											className="b2b-btn b2b-btn-secondary"
										>
											Edit Proforma
										</button>
									)}
									<button onClick={() => window.print()} className="b2b-btn b2b-btn-primary">Print / Download PDF</button>
								</div>
							</div>

							<div style={{ display: 'grid', gridTemplateColumns: '1fr 280px', gap: '24px' }}>
								{/* PROFORMA INVOICE DESIGN */}
								<div className="print-invoice-area" style={{ background: 'var(--surface-color)', padding: '40px', borderRadius: '16px', border: '1px solid var(--border-color)', color: 'var(--text-primary)', boxShadow: 'var(--shadow-sm)' }}>
									<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '36px' }}>
										<div>
											<h1 style={{ margin: '0 0 8px 0', textTransform: 'uppercase', color: 'var(--accent-color)', fontWeight: 800, fontSize: '2rem' }}>PROFORMA INVOICE</h1>
											<div style={{ fontSize: '14px', color: 'var(--text-secondary)' }}>
												<strong>Revision:</strong> v{selectedProforma.revision_number}
											</div>
										</div>
										<div style={{ textAlign: 'right' }}>
											<h2 style={{ margin: '0 0 6px 0', fontWeight: 800 }}>{selectedProforma.seller_name}</h2>
											<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedProforma.seller_gstin}</div>
											<div style={{ fontSize: '13px', color: 'var(--text-secondary)', maxWidth: '280px', lineHeight: 1.4 }}>{selectedProforma.seller_address}</div>
										</div>
									</div>

									<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', marginBottom: '24px' }} />

									<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1.2fr 1.1fr', gap: '30px', marginBottom: '36px' }}>
										<div>
											<h4 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontSize: '11px', letterSpacing: '0.05em' }}>Bill To</h4>
											<h3 style={{ margin: '0 0 6px 0', fontWeight: 700, fontSize: '15px' }}>{selectedProforma.customer_name}</h3>
											<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedProforma.customer_gstin}</div>
											<div style={{ fontSize: '13px', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', lineHeight: 1.4 }}>{selectedProforma.customer_address}</div>
											<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '4px' }}><strong>State:</strong> {selectedProforma.customer_state} ({selectedProforma.customer_state_code})</div>
										</div>
										<div>
											<h4 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontSize: '11px', letterSpacing: '0.05em' }}>Ship To</h4>
											<h3 style={{ margin: '0 0 6px 0', fontWeight: 700, fontSize: '15px' }}>{selectedProforma.customer_name}</h3>
											<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedProforma.customer_gstin}</div>
											<div style={{ fontSize: '13px', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', lineHeight: 1.4 }}>{selectedProforma.customer_shipping_address || selectedProforma.customer_address}</div>
											<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '4px' }}><strong>State:</strong> {selectedProforma.customer_state} ({selectedProforma.customer_state_code})</div>
										</div>
										<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1.5fr', gap: '10px 16px', fontSize: '13px', color: 'var(--text-secondary)' }}>
											<strong>Proforma Number:</strong>
											<span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{selectedProforma.proforma_number || 'DRAFT'}</span>

											<strong>Proforma Date:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedProforma.note_date ? selectedProforma.note_date.split('T')[0] : ''}</span>

											{selectedProforma.valid_until && (
												<>
													<strong>Valid Until:</strong>
													<span style={{ color: 'var(--text-primary)' }}>{selectedProforma.valid_until.split('T')[0]}</span>
												</>
											)}
										</div>
									</div>

									<table className="b2b-table" style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '36px' }}>
										<thead>
											<tr style={{ background: 'var(--bg-hover)', borderBottom: '2px solid var(--border-color)' }}>
												<th style={{ textAlign: 'left', padding: '12px 16px' }}>#</th>
												<th style={{ textAlign: 'left', padding: '12px 16px' }}>Product & Description</th>
												<th style={{ textAlign: 'center', padding: '12px 16px' }}>HSN</th>
												<th style={{ textAlign: 'right', padding: '12px 16px' }}>Qty</th>
												<th style={{ textAlign: 'right', padding: '12px 16px' }}>Rate (₹)</th>
												<th style={{ textAlign: 'right', padding: '12px 16px' }}>Amount (₹)</th>
											</tr>
										</thead>
										<tbody>
											{selectedProforma.items.map((item: any, i: number) => (
												<tr key={item.id || i} style={{ borderBottom: '1px solid var(--border-color)' }}>
													<td style={{ padding: '12px 16px', color: 'var(--text-tertiary)' }}>{i + 1}</td>
													<td style={{ padding: '12px 16px' }}>
														<div style={{ fontWeight: '500' }}>{item.item_details}</div>
														{item.sku && <div style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '2px' }}>SKU: {item.sku}</div>}
													</td>
													<td style={{ padding: '12px 16px', textAlign: 'center', fontFamily: 'monospace' }}>{item.hsn_code || '—'}</td>
													<td style={{ padding: '12px 16px', textAlign: 'right' }}>{item.quantity}</td>
													<td style={{ padding: '12px 16px', textAlign: 'right' }}>₹{item.rate.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
													<td style={{ padding: '12px 16px', textAlign: 'right', fontWeight: '500' }}>₹{item.amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</td>
												</tr>
											))}
										</tbody>
									</table>

									<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', alignItems: 'start' }}>
										<div style={{ fontSize: '12px', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
											<h4 style={{ margin: '0 0 6px 0', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontSize: '10px' }}>Notes</h4>
											This is a commercial Proforma Invoice containing estimated pricing and specifications. This is not a legal tax invoice. Legal GST compliance documents will be generated upon formal order acceptance and confirmation.
										</div>

										<div style={{ display: 'flex', flexDirection: 'column', gap: '10px', fontSize: '14px' }}>
											<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border-color)', paddingBottom: '6px' }}>
												<span>Subtotal:</span>
												<span>₹{selectedProforma.subtotal_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
											</div>

											{selectedProforma.discount_amount > 0 && (
												<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px solid var(--border-color)', paddingBottom: '6px', color: 'var(--danger-color)' }}>
													<span>Discount ({selectedProforma.discount_percent}%):</span>
													<span>-₹{selectedProforma.discount_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
												</div>
											)}

											{selectedProforma.cgst_amount > 0 && (
												<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px dashed var(--border-color)', paddingBottom: '6px', fontSize: '13px', color: 'var(--text-secondary)' }}>
													<span>CGST ({selectedProforma.cgst_rate}%):</span>
													<span>₹{selectedProforma.cgst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
												</div>
											)}

											{selectedProforma.sgst_amount > 0 && (
												<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px dashed var(--border-color)', paddingBottom: '6px', fontSize: '13px', color: 'var(--text-secondary)' }}>
													<span>SGST ({selectedProforma.sgst_rate}%):</span>
													<span>₹{selectedProforma.sgst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
												</div>
											)}

											{selectedProforma.igst_amount > 0 && (
												<div style={{ display: 'flex', justifyContent: 'space-between', borderBottom: '1px dashed var(--border-color)', paddingBottom: '6px', fontSize: '13px', color: 'var(--text-secondary)' }}>
													<span>IGST ({selectedProforma.igst_rate}%):</span>
													<span>₹{selectedProforma.igst_amount.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
												</div>
											)}

											<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '1.2rem', fontWeight: 800, color: 'var(--accent-color)', borderBottom: '2px solid var(--border-color)', paddingBottom: '8px' }}>
												<span>Total Amount:</span>
												<span>₹{selectedProforma.total_price.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
											</div>

											{selectedProforma.advance_paid > 0 && (
												<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '14px', color: 'var(--success-color)', fontWeight: '600' }}>
													<span>Advance Paid:</span>
													<span>₹{selectedProforma.advance_paid.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</span>
												</div>
											)}
										</div>
									</div>
								</div>

								{/* REVISION HISTORY STREAM (SIDEBAR) */}
								<div className="no-print" style={{ background: 'var(--bg-hover)', border: '1px solid var(--border-color)', borderRadius: '16px', padding: '20px' }}>
									<h4 style={{ margin: '0 0 16px 0', fontSize: '0.95rem', fontWeight: '700', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Negotiation History</h4>
									<div style={{ display: 'flex', flexDirection: 'column', gap: '16px', position: 'relative', paddingLeft: '12px', borderLeft: '2px solid var(--border-color)' }}>
										{historyStream.map(hist => (
											<div key={hist.id} style={{ position: 'relative' }}>
												{/* Marker dot */}
												<div style={{
													position: 'absolute',
													left: '-19px',
													top: '4px',
													width: '12px',
													height: '12px',
													borderRadius: '50%',
													background: hist.id === selectedProforma.id ? 'var(--accent-color)' : 'var(--text-tertiary)',
													border: '2px solid var(--surface-color)'
												}} />
												<div style={{ fontSize: '0.85rem' }}>
													<button
														onClick={() => setSelectedProforma(hist)}
														style={{
															background: 'none',
															border: 'none',
															padding: 0,
															textAlign: 'left',
															fontWeight: hist.id === selectedProforma.id ? '700' : '500',
															color: hist.id === selectedProforma.id ? 'var(--accent-color)' : 'var(--text-primary)',
															cursor: 'pointer',
															fontSize: '0.85rem'
														}}
													>
														Revision v{hist.revision_number} {hist.proforma_number ? `(${hist.proforma_number.split('-R')[0]})` : '(Draft)'}
													</button>
													<div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '2px' }}>
														{new Date(hist.note_date).toLocaleDateString('en-IN', { day: '2-digit', month: 'short' })} • <span style={{ textTransform: 'capitalize' }}>{hist.status.toLowerCase()}</span>
													</div>
												</div>
											</div>
										))}
									</div>
								</div>
							</div>
						</div>
					);
				})()}


				{/* CREDIT NOTE CREATOR & EDITOR VIEW */}
				{(viewMode === 'create-cn' || viewMode === 'edit-cn') && (
					<div className="b2b-form-area no-print" style={{ color: 'var(--text-primary)' }}>
						<div className="form-header-container">
							<div>
								<h3 className="form-header-title">
									<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ color: 'var(--accent-color)' }}>
										<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
										<polyline points="14 2 14 8 20 8" />
									</svg>
									{viewMode === 'create-cn' ? 'Create Credit Note' : 'Edit Credit Note'}
								</h3>
								<p className="form-header-subtitle">Issue a credit note to reduce client outstanding balance.</p>
							</div>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary" style={{ height: '40px' }}>
								&larr; Back to List
							</button>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Link Invoice & Client Info</div>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', alignItems: 'start' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
									<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
										<label className="form-label">Link B2B Invoice (Optional)</label>
										<select
											className="b2b-input"
											value={formCreditNote.invoice_id || ''}
											onChange={(e) => {
												const invId = Number(e.target.value);
												const inv = invoices.find(i => i.id === invId);
												if (inv) {
													const updated = {
														...formCreditNote,
														invoice_id: inv.id,
														invoice_number: inv.invoice_number,
														customer_id: inv.customer_id,
														customer_gstin: inv.customer_gstin,
														customer_name: inv.customer_name,
														customer_state: inv.customer_state,
														customer_state_code: inv.customer_state_code,
														customer_address: inv.customer_address,
														items: inv.items.map(item => ({
															product_id: item.product_id,
															item_details: item.item_details,
															sku: item.sku,
															hsn_code: item.hsn_code,
															quantity: item.quantity,
															rate: item.rate,
															amount: item.amount
														}))
													};
													recalculateCreditNoteTotals(updated);
												}
											}}
											style={{ height: '46px' }}
										>
											<option value="">-- Select Invoice --</option>
											{invoices.filter(i => i.status === 'ISSUED').map(i => (
												<option key={i.id} value={i.id}>{i.invoice_number} - {i.customer_name} (₹{i.total_price.toFixed(2)})</option>
											))}
										</select>
									</div>

									<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
										<label className="form-label">Reason for Credit Note*</label>
										<input
											type="text"
											className="b2b-input"
											placeholder="e.g. Sales return / Damaged items / Quantity correction"
											value={formCreditNote.reason || ''}
											onChange={(e) => setFormCreditNote({ ...formCreditNote, reason: e.target.value })}
										/>
									</div>
								</div>

								{formCreditNote.customer_gstin ? (
									<div className="client-info-card">
										<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
											<strong style={{ fontSize: '1rem', color: 'var(--text-primary)' }}>{formCreditNote.customer_name}</strong>
											<span className="client-badge">GST ACTIVE</span>
										</div>
										<div style={{ display: 'grid', gridTemplateColumns: '80px 1fr', gap: '6px', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
											<span><strong>GSTIN:</strong></span>
											<span style={{ color: 'var(--text-primary)', fontFamily: 'monospace', fontWeight: 600 }}>{formCreditNote.customer_gstin}</span>
											<span><strong>State:</strong></span>
											<span>{formCreditNote.customer_state} ({formCreditNote.customer_state_code})</span>
											<span><strong>Address:</strong></span>
											<span style={{ textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }} title={formCreditNote.customer_address}>{formCreditNote.customer_address}</span>
										</div>
									</div>
								) : (
									<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '120px', border: '1px dashed var(--border-color)', borderRadius: '12px', background: 'var(--bg-hover)', color: 'var(--text-tertiary)', fontSize: '0.85rem' }}>
										Select a linked B2B invoice to automatically prefill client information and line items.
									</div>
								)}
							</div>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title font-bold">Credit Note Items</div>
							<div className="table-responsive" style={{ overflowX: 'auto', marginBottom: '20px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
								<table style={{ minWidth: '900px', width: '100%', borderCollapse: 'collapse' }}>
									<thead>
										<tr className="items-table-header" style={{ borderBottom: '2px solid var(--border-color)', textAlign: 'left', background: 'var(--bg-hover)' }}>
											<th style={{ padding: '14px 16px', color: 'var(--text-secondary)' }}>Item Details</th>
											<th style={{ padding: '14px 16px', color: 'var(--text-secondary)' }}>Quantity</th>
											<th style={{ padding: '14px 16px', color: 'var(--text-secondary)' }}>Rate (₹)</th>
											<th style={{ padding: '14px 16px', color: 'var(--text-secondary)' }}>Amount</th>
											<th style={{ padding: '14px 16px', textAlign: 'right' }}></th>
										</tr>
									</thead>
									<tbody>
										{formCreditNote.items.map((item: any, idx: number) => (
											<tr key={idx} className="items-table-row" style={{ borderBottom: '1px solid var(--border-color)' }}>
												<td style={{ padding: '12px 16px' }}>
													<input
														type="text"
														className="b2b-input"
														placeholder="Line item details..."
														value={item.item_details}
														onChange={(e) => {
															const newItems = [...formCreditNote.items];
															newItems[idx].item_details = e.target.value;
															recalculateCreditNoteTotals({ ...formCreditNote, items: newItems });
														}}
													/>
												</td>
												<td style={{ padding: '12px 16px', width: '140px' }}>
													<input
														type="number"
														className="b2b-input"
														value={item.quantity}
														min="1"
														onChange={(e) => {
															const newItems = [...formCreditNote.items];
															newItems[idx].quantity = Number(e.target.value);
															recalculateCreditNoteTotals({ ...formCreditNote, items: newItems });
														}}
													/>
												</td>
												<td style={{ padding: '12px 16px', width: '160px' }}>
													<input
														type="number"
														className="b2b-input"
														value={item.rate}
														onChange={(e) => {
															const newItems = [...formCreditNote.items];
															newItems[idx].rate = Number(e.target.value);
															recalculateCreditNoteTotals({ ...formCreditNote, items: newItems });
														}}
													/>
												</td>
												<td style={{ padding: '12px 16px', fontWeight: 'bold' }}>
													₹{(item.quantity * item.rate).toFixed(2)}
												</td>
												<td style={{ padding: '12px 16px', textAlign: 'right' }}>
													{formCreditNote.items.length > 1 && (
														<button
															onClick={() => {
																const newItems = formCreditNote.items.filter((_: any, i: number) => i !== idx);
																recalculateCreditNoteTotals({ ...formCreditNote, items: newItems });
															}}
															className="delete-row-btn"
														>
															✕
														</button>
													)}
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
							<button
								onClick={() => {
									const newItems = [...formCreditNote.items, { item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }];
									setFormCreditNote({ ...formCreditNote, items: newItems });
								}}
								className="add-row-btn"
							>
								+ Add Row
							</button>
						</div>

						<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', marginTop: '24px' }}>
							<div>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Note Date*</label>
									<input
										type="date"
										className="b2b-input"
										value={formCreditNote.note_date}
										onChange={(e) => setFormCreditNote({ ...formCreditNote, note_date: e.target.value })}
									/>
								</div>
							</div>
							<div className="summary-panel">
								<div className="summary-row">
									<span>Sub Total:</span>
									<strong>₹{formCreditNote.subtotal_price.toFixed(2)}</strong>
								</div>
								{formCreditNote.cgst_amount > 0 && (
									<div className="summary-row">
										<span>CGST ({formCreditNote.cgst_rate}%):</span>
										<span>₹{formCreditNote.cgst_amount.toFixed(2)}</span>
									</div>
								)}
								{formCreditNote.sgst_amount > 0 && (
									<div className="summary-row">
										<span>SGST ({formCreditNote.sgst_rate}%):</span>
										<span>₹{formCreditNote.sgst_amount.toFixed(2)}</span>
									</div>
								)}
								{formCreditNote.igst_amount > 0 && (
									<div className="summary-row">
										<span>IGST ({formCreditNote.igst_rate}%):</span>
										<span>₹{formCreditNote.igst_amount.toFixed(2)}</span>
									</div>
								)}
								<div className="summary-total-box">
									<span>Total Credit Amount:</span>
									<strong style={{ fontSize: '1.3rem', color: 'var(--accent-color)' }}>₹{formCreditNote.total_price.toFixed(2)}</strong>
								</div>
							</div>
						</div>

						<div style={{ display: 'flex', justifyContent: 'flex-end', gap: '16px', marginTop: '40px', borderTop: '1px solid var(--border-color)', paddingTop: '28px' }}>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary">Cancel</button>
							<button onClick={() => handleSaveCreditNote(true)} className="b2b-btn b2b-btn-secondary">Save as Draft</button>
							<button onClick={() => handleSaveCreditNote(false)} className="b2b-btn b2b-btn-primary">Save & Issue Note</button>
						</div>
					</div>
				)}

				{/* DEBIT NOTE CREATOR & EDITOR VIEW */}
				{(viewMode === 'create-dn' || viewMode === 'edit-dn') && (
					<div className="b2b-form-area no-print" style={{ color: 'var(--text-primary)' }}>
						<div className="form-header-container">
							<div>
								<h3 className="form-header-title">
									<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ color: 'var(--accent-color)' }}>
										<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
										<polyline points="14 2 14 8 20 8" />
									</svg>
									{viewMode === 'create-dn' ? 'Create Debit Note' : 'Edit Debit Note'}
								</h3>
								<p className="form-header-subtitle">Issue a debit note to increase client outstanding balance.</p>
							</div>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary" style={{ height: '40px' }}>
								&larr; Back to List
							</button>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title">Link Invoice & Client Info</div>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', alignItems: 'start' }}>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
									<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
										<label className="form-label">Link B2B Invoice (Optional)</label>
										<select
											className="b2b-input"
											value={formDebitNote.invoice_id || ''}
											onChange={(e) => {
												const invId = Number(e.target.value);
												const inv = invoices.find(i => i.id === invId);
												if (inv) {
													const updated = {
														...formDebitNote,
														invoice_id: inv.id,
														invoice_number: inv.invoice_number,
														customer_id: inv.customer_id,
														customer_gstin: inv.customer_gstin,
														customer_name: inv.customer_name,
														customer_state: inv.customer_state,
														customer_state_code: inv.customer_state_code,
														customer_address: inv.customer_address,
														items: inv.items.map(item => ({
															product_id: item.product_id,
															item_details: item.item_details,
															sku: item.sku,
															hsn_code: item.hsn_code,
															quantity: item.quantity,
															rate: item.rate,
															amount: item.amount
														}))
													};
													recalculateDebitNoteTotals(updated);
												}
											}}
											style={{ height: '46px' }}
										>
											<option value="">-- Select Invoice --</option>
											{invoices.filter(i => i.status === 'ISSUED').map(i => (
												<option key={i.id} value={i.id}>{i.invoice_number} - {i.customer_name} (₹{i.total_price.toFixed(2)})</option>
											))}
										</select>
									</div>

									<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
										<label className="form-label">Reason for Debit Note*</label>
										<input
											type="text"
											className="b2b-input"
											placeholder="e.g. Value correction / Under-billing adjustment"
											value={formDebitNote.reason || ''}
											onChange={(e) => setFormDebitNote({ ...formDebitNote, reason: e.target.value })}
										/>
									</div>
								</div>

								{formDebitNote.customer_gstin ? (
									<div className="client-info-card">
										<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
											<strong style={{ fontSize: '1rem', color: 'var(--text-primary)' }}>{formDebitNote.customer_name}</strong>
											<span className="client-badge">GST ACTIVE</span>
										</div>
										<div style={{ display: 'grid', gridTemplateColumns: '80px 1fr', gap: '6px', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
											<span><strong>GSTIN:</strong></span>
											<span style={{ color: 'var(--text-primary)', fontFamily: 'monospace', fontWeight: 600 }}>{formDebitNote.customer_gstin}</span>
											<span><strong>State:</strong></span>
											<span>{formDebitNote.customer_state} ({formDebitNote.customer_state_code})</span>
											<span><strong>Address:</strong></span>
											<span style={{ textOverflow: 'ellipsis', overflow: 'hidden', whiteSpace: 'nowrap' }} title={formDebitNote.customer_address}>{formDebitNote.customer_address}</span>
										</div>
									</div>
								) : (
									<div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '120px', border: '1px dashed var(--border-color)', borderRadius: '12px', background: 'var(--bg-hover)', color: 'var(--text-tertiary)', fontSize: '0.85rem' }}>
										Select a linked B2B invoice to automatically prefill client information and line items.
									</div>
								)}
							</div>
						</div>

						<div className="b2b-form-section">
							<div className="b2b-form-section-title font-bold">Debit Note Items</div>
							<div className="table-responsive" style={{ overflowX: 'auto', marginBottom: '20px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
								<table style={{ minWidth: '900px', width: '100%', borderCollapse: 'collapse' }}>
									<thead>
										<tr className="items-table-header" style={{ borderBottom: '2px solid var(--border-color)', textAlign: 'left', background: 'var(--bg-hover)' }}>
											<th style={{ padding: '14px 16px', color: 'var(--text-secondary)' }}>Item Details</th>
											<th style={{ padding: '14px 16px', color: 'var(--text-secondary)' }}>Quantity</th>
											<th style={{ padding: '14px 16px', color: 'var(--text-secondary)' }}>Rate (₹)</th>
											<th style={{ padding: '14px 16px', color: 'var(--text-secondary)' }}>Amount</th>
											<th style={{ padding: '14px 16px', textAlign: 'right' }}></th>
										</tr>
									</thead>
									<tbody>
										{formDebitNote.items.map((item: any, idx: number) => (
											<tr key={idx} className="items-table-row" style={{ borderBottom: '1px solid var(--border-color)' }}>
												<td style={{ padding: '12px 16px' }}>
													<input
														type="text"
														className="b2b-input"
														placeholder="Line item details..."
														value={item.item_details}
														onChange={(e) => {
															const newItems = [...formDebitNote.items];
															newItems[idx].item_details = e.target.value;
															recalculateDebitNoteTotals({ ...formDebitNote, items: newItems });
														}}
													/>
												</td>
												<td style={{ padding: '12px 16px', width: '140px' }}>
													<input
														type="number"
														className="b2b-input"
														value={item.quantity}
														min="1"
														onChange={(e) => {
															const newItems = [...formDebitNote.items];
															newItems[idx].quantity = Number(e.target.value);
															recalculateDebitNoteTotals({ ...formDebitNote, items: newItems });
														}}
													/>
												</td>
												<td style={{ padding: '12px 16px', width: '160px' }}>
													<input
														type="number"
														className="b2b-input"
														value={item.rate}
														onChange={(e) => {
															const newItems = [...formDebitNote.items];
															newItems[idx].rate = Number(e.target.value);
															recalculateDebitNoteTotals({ ...formDebitNote, items: newItems });
														}}
													/>
												</td>
												<td style={{ padding: '12px 16px', fontWeight: 'bold' }}>
													₹{(item.quantity * item.rate).toFixed(2)}
												</td>
												<td style={{ padding: '12px 16px', textAlign: 'right' }}>
													{formDebitNote.items.length > 1 && (
														<button
															onClick={() => {
																const newItems = formDebitNote.items.filter((_: any, i: number) => i !== idx);
																recalculateDebitNoteTotals({ ...formDebitNote, items: newItems });
															}}
															className="delete-row-btn"
														>
															✕
														</button>
													)}
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
							<button
								onClick={() => {
									const newItems = [...formDebitNote.items, { item_details: '', quantity: 1, rate: 0, amount: 0, hsn_code: '33029019' }];
									setFormDebitNote({ ...formDebitNote, items: newItems });
								}}
								className="add-row-btn"
							>
								+ Add Row
							</button>
						</div>

						<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', marginTop: '24px' }}>
							<div>
								<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
									<label className="form-label">Note Date*</label>
									<input
										type="date"
										className="b2b-input"
										value={formDebitNote.note_date}
										onChange={(e) => setFormDebitNote({ ...formDebitNote, note_date: e.target.value })}
									/>
								</div>
							</div>
							<div className="summary-panel">
								<div className="summary-row">
									<span>Sub Total:</span>
									<strong>₹{formDebitNote.subtotal_price.toFixed(2)}</strong>
								</div>
								{formDebitNote.cgst_amount > 0 && (
									<div className="summary-row">
										<span>CGST ({formDebitNote.cgst_rate}%):</span>
										<span>₹{formDebitNote.cgst_amount.toFixed(2)}</span>
									</div>
								)}
								{formDebitNote.sgst_amount > 0 && (
									<div className="summary-row">
										<span>SGST ({formDebitNote.sgst_rate}%):</span>
										<span>₹{formDebitNote.sgst_amount.toFixed(2)}</span>
									</div>
								)}
								{formDebitNote.igst_amount > 0 && (
									<div className="summary-row">
										<span>IGST ({formDebitNote.igst_rate}%):</span>
										<span>₹{formDebitNote.igst_amount.toFixed(2)}</span>
									</div>
								)}
								<div className="summary-total-box">
									<span>Total Debit Amount:</span>
									<strong style={{ fontSize: '1.3rem', color: 'var(--accent-color)' }}>₹{formDebitNote.total_price.toFixed(2)}</strong>
								</div>
							</div>
						</div>

						<div style={{ display: 'flex', justifyContent: 'flex-end', gap: '16px', marginTop: '40px', borderTop: '1px solid var(--border-color)', paddingTop: '28px' }}>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary">Cancel</button>
							<button onClick={() => handleSaveDebitNote(true)} className="b2b-btn b2b-btn-secondary">Save as Draft</button>
							<button onClick={() => handleSaveDebitNote(false)} className="b2b-btn b2b-btn-primary">Save & Issue Note</button>
						</div>
					</div>
				)}

				{/* CREDIT NOTE PREVIEW */}
				{viewMode === 'preview-cn' && selectedCreditNote && (
					<div>
						<div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '24px', borderBottom: '1px solid var(--border-color)', paddingBottom: '12px', alignItems: 'center' }}>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary">&larr; Back to List</button>
							<div style={{ display: 'flex', gap: '8px' }}>
								{selectedCreditNote.status === 'DRAFT' && userRole === 'admin' && (
									<button
										onClick={() => {
											setFormCreditNote({ ...selectedCreditNote });
											setViewMode('edit-cn');
										}}
										className="b2b-btn b2b-btn-secondary"
									>
										Edit Credit Note
									</button>
								)}
								<button onClick={triggerPrint} className="b2b-btn b2b-btn-primary">Print / Download Credit Note</button>
							</div>
						</div>

						<div className="print-invoice-area" style={{ background: 'var(--surface-color)', padding: '40px', borderRadius: '16px', border: '1px solid var(--border-color)', color: 'var(--text-primary)', boxShadow: 'var(--shadow-sm)' }}>
							<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '36px' }}>
								<div>
									<h1 style={{ margin: '0 0 8px 0', textTransform: 'uppercase', color: 'var(--accent-color)', fontWeight: 800, fontSize: '2rem' }}>CREDIT NOTE</h1>
									<div style={{ fontSize: '14px', color: 'var(--text-secondary)' }}><strong>FY:</strong> {selectedCreditNote.financial_year || 'N/A'}</div>
								</div>
								<div style={{ textAlign: 'right' }}>
									<h2 style={{ margin: '0 0 6px 0', fontWeight: 800 }}>{selectedCreditNote.seller_name}</h2>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedCreditNote.seller_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', maxWidth: '280px', lineHeight: 1.4 }}>{selectedCreditNote.seller_address}</div>
								</div>
							</div>

							<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', marginBottom: '24px' }} />

							<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', marginBottom: '36px' }}>
								<div>
									<h4 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontSize: '11px', letterSpacing: '0.05em' }}>Issued To</h4>
									<h3 style={{ margin: '0 0 6px 0', fontWeight: 700 }}>{selectedCreditNote.customer_name}</h3>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedCreditNote.customer_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', lineHeight: 1.4 }}>{selectedCreditNote.customer_address}</div>
								</div>
								<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1.5fr', gap: '10px 16px', fontSize: '13px', color: 'var(--text-secondary)' }}>
									<strong>Credit Note Number:</strong>
									<span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{selectedCreditNote.credit_note_number || 'DRAFT'}</span>

									<strong>Note Date:</strong>
									<span style={{ color: 'var(--text-primary)' }}>{selectedCreditNote.note_date ? selectedCreditNote.note_date.split('T')[0] : ''}</span>

									{selectedCreditNote.invoice_number && (
										<>
											<strong>Linked Invoice:</strong>
											<span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{selectedCreditNote.invoice_number}</span>
										</>
									)}

									{selectedCreditNote.reason && (
										<>
											<strong>Reason:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedCreditNote.reason}</span>
										</>
									)}
								</div>
							</div>

							<table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '36px' }}>
								<thead>
									<tr style={{ borderBottom: '2px solid var(--border-color)', textAlign: 'left', fontSize: '13px', color: 'var(--text-secondary)' }}>
										<th style={{ padding: '12px 8px', fontWeight: 600 }}>Item Details</th>
										<th style={{ padding: '12px 8px', width: '100px', textAlign: 'right', fontWeight: 600 }}>Qty</th>
										<th style={{ padding: '12px 8px', width: '120px', textAlign: 'right', fontWeight: 600 }}>Rate (₹)</th>
										<th style={{ padding: '12px 8px', width: '120px', textAlign: 'right', fontWeight: 600 }}>Amount (₹)</th>
									</tr>
								</thead>
								<tbody>
									{selectedCreditNote.items && selectedCreditNote.items.map((item: any, idx: number) => (
										<tr key={idx} style={{ borderBottom: '1px solid var(--border-color)', fontSize: '14px', color: 'var(--text-primary)' }}>
											<td style={{ padding: '12px 8px' }}>
												<strong>{item.item_details}</strong>
											</td>
											<td style={{ padding: '12px 8px', textAlign: 'right' }}>{item.quantity}</td>
											<td style={{ padding: '12px 8px', textAlign: 'right' }}>₹{item.rate.toFixed(2)}</td>
											<td style={{ padding: '12px 8px', textAlign: 'right', fontWeight: 'bold' }}>₹{item.amount.toFixed(2)}</td>
										</tr>
									))}
								</tbody>
							</table>

							<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px' }}>
								<div></div>
								<div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
									<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
										<span>Sub Total:</span>
										<span style={{ color: 'var(--text-primary)' }}>₹{selectedCreditNote.subtotal_price.toFixed(2)}</span>
									</div>
									{selectedCreditNote.cgst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>CGST ({selectedCreditNote.cgst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedCreditNote.cgst_amount.toFixed(2)}</span>
										</div>
									)}
									{selectedCreditNote.sgst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>SGST ({selectedCreditNote.sgst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedCreditNote.sgst_amount.toFixed(2)}</span>
										</div>
									)}
									{selectedCreditNote.igst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>IGST ({selectedCreditNote.igst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedCreditNote.igst_amount.toFixed(2)}</span>
										</div>
									)}
									<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', margin: '12px 0' }} />
									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '18px', fontWeight: 800, color: 'var(--text-primary)' }}>
										<span>Total Credit (₹):</span>
										<span>₹{selectedCreditNote.total_price.toFixed(2)}</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				)}

				{/* DEBIT NOTE PREVIEW */}
				{viewMode === 'preview-dn' && selectedDebitNote && (
					<div>
						<div className="no-print" style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '24px', borderBottom: '1px solid var(--border-color)', paddingBottom: '12px', alignItems: 'center' }}>
							<button onClick={() => setViewMode('list')} className="b2b-btn b2b-btn-secondary">&larr; Back to List</button>
							<div style={{ display: 'flex', gap: '8px' }}>
								{selectedDebitNote.status === 'DRAFT' && userRole === 'admin' && (
									<button
										onClick={() => {
											setFormDebitNote({ ...selectedDebitNote });
											setViewMode('edit-dn');
										}}
										className="b2b-btn b2b-btn-secondary"
									>
										Edit Debit Note
									</button>
								)}
								<button onClick={triggerPrint} className="b2b-btn b2b-btn-primary">Print / Download Debit Note</button>
							</div>
						</div>

						<div className="print-invoice-area" style={{ background: 'var(--surface-color)', padding: '40px', borderRadius: '16px', border: '1px solid var(--border-color)', color: 'var(--text-primary)', boxShadow: 'var(--shadow-sm)' }}>
							<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '36px' }}>
								<div>
									<h1 style={{ margin: '0 0 8px 0', textTransform: 'uppercase', color: 'var(--accent-color)', fontWeight: 800, fontSize: '2rem' }}>DEBIT NOTE</h1>
									<div style={{ fontSize: '14px', color: 'var(--text-secondary)' }}><strong>FY:</strong> {selectedDebitNote.financial_year || 'N/A'}</div>
								</div>
								<div style={{ textAlign: 'right' }}>
									<h2 style={{ margin: '0 0 6px 0', fontWeight: 800 }}>{selectedDebitNote.seller_name}</h2>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedDebitNote.seller_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', maxWidth: '280px', lineHeight: 1.4 }}>{selectedDebitNote.seller_address}</div>
								</div>
							</div>

							<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', marginBottom: '24px' }} />

							<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px', marginBottom: '36px' }}>
								<div>
									<h4 style={{ margin: '0 0 8px 0', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontSize: '11px', letterSpacing: '0.05em' }}>Issued To</h4>
									<h3 style={{ margin: '0 0 6px 0', fontWeight: 700 }}>{selectedDebitNote.customer_name}</h3>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '4px' }}><strong>GSTIN:</strong> {selectedDebitNote.customer_gstin}</div>
									<div style={{ fontSize: '13px', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', lineHeight: 1.4 }}>{selectedDebitNote.customer_address}</div>
								</div>
								<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1.5fr', gap: '10px 16px', fontSize: '13px', color: 'var(--text-secondary)' }}>
									<strong>Debit Note Number:</strong>
									<span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{selectedDebitNote.debit_note_number || 'DRAFT'}</span>

									<strong>Note Date:</strong>
									<span style={{ color: 'var(--text-primary)' }}>{selectedDebitNote.note_date ? selectedDebitNote.note_date.split('T')[0] : ''}</span>

									{selectedDebitNote.invoice_number && (
										<>
											<strong>Linked Invoice:</strong>
											<span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{selectedDebitNote.invoice_number}</span>
										</>
									)}

									{selectedDebitNote.reason && (
										<>
											<strong>Reason:</strong>
											<span style={{ color: 'var(--text-primary)' }}>{selectedDebitNote.reason}</span>
										</>
									)}
								</div>
							</div>

							<table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '36px' }}>
								<thead>
									<tr style={{ borderBottom: '2px solid var(--border-color)', textAlign: 'left', fontSize: '13px', color: 'var(--text-secondary)' }}>
										<th style={{ padding: '12px 8px', fontWeight: 600 }}>Item Details</th>
										<th style={{ padding: '12px 8px', width: '100px', textAlign: 'right', fontWeight: 600 }}>Qty</th>
										<th style={{ padding: '12px 8px', width: '120px', textAlign: 'right', fontWeight: 600 }}>Rate (₹)</th>
										<th style={{ padding: '12px 8px', width: '120px', textAlign: 'right', fontWeight: 600 }}>Amount (₹)</th>
									</tr>
								</thead>
								<tbody>
									{selectedDebitNote.items && selectedDebitNote.items.map((item: any, idx: number) => (
										<tr key={idx} style={{ borderBottom: '1px solid var(--border-color)', fontSize: '14px', color: 'var(--text-primary)' }}>
											<td style={{ padding: '12px 8px' }}>
												<strong>{item.item_details}</strong>
											</td>
											<td style={{ padding: '12px 8px', textAlign: 'right' }}>{item.quantity}</td>
											<td style={{ padding: '12px 8px', textAlign: 'right' }}>₹{item.rate.toFixed(2)}</td>
											<td style={{ padding: '12px 8px', textAlign: 'right', fontWeight: 'bold' }}>₹{item.amount.toFixed(2)}</td>
										</tr>
									))}
								</tbody>
							</table>

							<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '40px' }}>
								<div></div>
								<div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
									<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
										<span>Sub Total:</span>
										<span style={{ color: 'var(--text-primary)' }}>₹{selectedDebitNote.subtotal_price.toFixed(2)}</span>
									</div>
									{selectedDebitNote.cgst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>CGST ({selectedDebitNote.cgst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedDebitNote.cgst_amount.toFixed(2)}</span>
										</div>
									)}
									{selectedDebitNote.sgst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>SGST ({selectedDebitNote.sgst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedDebitNote.sgst_amount.toFixed(2)}</span>
										</div>
									)}
									{selectedDebitNote.igst_amount > 0 && (
										<div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
											<span>IGST ({selectedDebitNote.igst_rate}%):</span>
											<span style={{ color: 'var(--text-primary)' }}>₹{selectedDebitNote.igst_amount.toFixed(2)}</span>
										</div>
									)}
									<hr style={{ border: 'none', borderBottom: '1px solid var(--border-color)', margin: '12px 0' }} />
									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '18px', fontWeight: 800, color: 'var(--text-primary)' }}>
										<span>Total Debit (₹):</span>
										<span>₹{selectedDebitNote.total_price.toFixed(2)}</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				)}
			</div>

			{/* CUSTOMER REGISTRATION DIALOG MODAL */}
			{showCustomerModal && editingCustomer && createPortal(
				<div className="modal-overlay" onClick={() => {
					setShowCustomerModal(false);
					setEditingCustomer(null);
				}} style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'center', overflowY: 'auto', padding: '2rem 1.5rem', zIndex: 3000 }}>
					<div className="premium-modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '800px', width: '100%', margin: '0 auto', border: '1px solid var(--border-color)', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.15)', borderRadius: '24px', padding: '2.5rem', background: 'var(--surface-color)', position: 'relative' }}>

						{/* Modal Header */}
						<div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '2rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1.25rem' }}>
							<div style={{
								width: '48px',
								height: '48px',
								borderRadius: '12px',
								background: 'linear-gradient(135deg, var(--accent-color), #10b981)',
								display: 'flex',
								alignItems: 'center',
								justifyContent: 'center',
								color: 'white',
								boxShadow: '0 8px 16px rgba(16, 185, 129, 0.15)',
								flexShrink: 0
							}}>
								<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<path d="M17 21v-2a4 4 0 0 0-3-3.87" />
									<path d="M16 3.13a4 4 0 0 1 0 7.75" />
									<circle cx="9" cy="7" r="4" />
									<path d="M17 11a5.5 5.5 0 0 0-4.5 4.5" />
									<path d="M2 21v-2a4 4 0 0 1 4-4h6a4 4 0 0 1 4 4v2" />
								</svg>
							</div>
							<div>
								<h2 style={{ margin: 0, fontSize: '1.45rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>
									{editingCustomer.id ? 'Edit Client Details' : 'Register New Client'}
								</h2>
								<p style={{ margin: '3px 0 0 0', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
									Provide the official GSTIN, billing address, and shipping address details.
								</p>
							</div>
						</div>

						<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px' }}>
							{/* Left Column: Business Details */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
								<div className="sync-form-group">
									<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Legal Business Name*</label>
									<input
										type="text"
										className="b2b-input"
										value={editingCustomer.legal_name}
										onChange={(e) => setEditingCustomer({ ...editingCustomer, legal_name: e.target.value })}
										placeholder="e.g. Acme Corporation Pvt Ltd"
									/>
								</div>

								<div className="sync-form-group">
									<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Trade Name (Optional)</label>
									<input
										type="text"
										className="b2b-input"
										value={editingCustomer.trade_name || ''}
										onChange={(e) => setEditingCustomer({ ...editingCustomer, trade_name: e.target.value })}
										placeholder="e.g. Acme Stores"
									/>
								</div>

								<div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: '16px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Client GSTIN*</label>
										<input
											type="text"
											placeholder="e.g. 33ABCDE1234F1Z5"
											maxLength={15}
											className="b2b-input"
											value={editingCustomer.gstin}
											onChange={(e) => {
												const val = e.target.value.toUpperCase();
												const updated = { ...editingCustomer, gstin: val };
												if (val.length >= 2) {
													const code = val.substring(0, 2);
													const detectedState = GST_STATE_MAP[code];
													if (detectedState) {
														updated.state_code = code;
														updated.state = detectedState;
													}
												}
												if (val.length >= 12) {
													const pan = val.substring(2, 12);
													updated.pan = pan;
												}
												setEditingCustomer(updated);
											}}
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>PAN</label>
										<input
											type="text"
											placeholder="Auto-extracted"
											maxLength={10}
											className="b2b-input"
											value={editingCustomer.pan || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, pan: e.target.value.toUpperCase() })}
										/>
									</div>
								</div>
							</div>

							{/* Right Column: Contact & GST Treatment */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
								<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Email</label>
										<input
											type="email"
											className="b2b-input"
											value={editingCustomer.email || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, email: e.target.value })}
											placeholder="billing@acme.com"
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Phone</label>
										<input
											type="text"
											className="b2b-input"
											value={editingCustomer.phone || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, phone: e.target.value })}
											placeholder="e.g. +91 9876543210"
										/>
									</div>
								</div>

								<div className="sync-form-group">
									<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>GST Treatment</label>
									<select className="b2b-input" style={{ appearance: 'auto' }} defaultValue="regular">
										<option value="regular">Registered Business - Regular</option>
										<option value="composition">Registered Business - Composition</option>
										<option value="unregistered">Unregistered Business</option>
										<option value="consumer">Consumer</option>
										<option value="overseas">Overseas</option>
									</select>
								</div>

								<div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: '16px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>State / Place of Supply</label>
										<input
											type="text"
											className="b2b-input"
											value={editingCustomer.state || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, state: e.target.value })}
											placeholder="e.g. Tamil Nadu"
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>State Code</label>
										<input
											type="text"
											className="b2b-input"
											value={editingCustomer.state_code || ''}
											onChange={(e) => setEditingCustomer({ ...editingCustomer, state_code: e.target.value })}
											placeholder="e.g. 33"
											maxLength={2}
										/>
									</div>
								</div>
							</div>
						</div>

						{/* Addresses Section */}
						<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '24px', borderTop: '1px solid var(--border-color)', marginTop: '24px', paddingTop: '20px' }}>
							{/* Billing Address Card */}
							<div style={{ background: 'var(--bg-input)', padding: '16px', borderRadius: '16px', border: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '12px' }}>
								<h3 style={{ margin: '0 0 4px 0', fontSize: '0.8rem', fontWeight: 800, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'flex', alignItems: 'center', gap: '6px' }}>
									<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ color: 'var(--accent-color)' }}>
										<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
										<circle cx="12" cy="10" r="3" />
									</svg>
									Billing Address
								</h3>

								<div className="sync-form-group">
									<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>Street Address</label>
									<input
										type="text"
										className="b2b-input"
										value={billingStreet}
										onChange={(e) => setBillingStreet(e.target.value)}
										placeholder="Street, Area, Building"
									/>
								</div>

								<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>City</label>
										<input
											type="text"
											className="b2b-input"
											value={billingCity}
											onChange={(e) => setBillingCity(e.target.value)}
											placeholder="e.g. Chennai"
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>Pin Code</label>
										<input
											type="text"
											className="b2b-input"
											value={billingPincode}
											onChange={(e) => setBillingPincode(e.target.value)}
											placeholder="6-digit PIN"
											maxLength={6}
										/>
									</div>
								</div>
							</div>

							{/* Shipping Address Card */}
							<div style={{ background: 'var(--bg-input)', padding: '16px', borderRadius: '16px', border: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '12px', opacity: sameAsBilling ? 0.85 : 1, transition: 'opacity 0.2s ease' }}>
								<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '4px' }}>
									<h3 style={{ margin: 0, fontSize: '0.8rem', fontWeight: 800, color: 'var(--text-primary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'flex', alignItems: 'center', gap: '6px' }}>
										<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" style={{ color: sameAsBilling ? 'var(--text-tertiary)' : 'var(--accent-color)' }}>
											<rect width="16" height="13" x="2" y="6" rx="2" />
											<path d="M16 2h4l3 4v13a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h4" />
										</svg>
										Shipping Address
									</h3>
									<div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
										<input
											type="checkbox"
											id="sameAsBilling"
											checked={sameAsBilling}
											onChange={(e) => setSameAsBilling(e.target.checked)}
											style={{ width: 'auto', cursor: 'pointer' }}
										/>
										<label htmlFor="sameAsBilling" style={{ fontWeight: 700, fontSize: '0.75rem', cursor: 'pointer', color: 'var(--text-secondary)' }}>Same as Billing</label>
									</div>
								</div>

								<div className="sync-form-group">
									<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>Street Address</label>
									<input
										type="text"
										className="b2b-input"
										value={sameAsBilling ? billingStreet : shippingStreet}
										onChange={(e) => setShippingStreet(e.target.value)}
										placeholder="Street, Area, Building"
										disabled={sameAsBilling}
										style={{ opacity: sameAsBilling ? 0.6 : 1 }}
									/>
								</div>

								<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>City</label>
										<input
											type="text"
											className="b2b-input"
											value={sameAsBilling ? billingCity : shippingCity}
											onChange={(e) => setShippingCity(e.target.value)}
											placeholder="e.g. Chennai"
											disabled={sameAsBilling}
											style={{ opacity: sameAsBilling ? 0.6 : 1 }}
										/>
									</div>
									<div className="sync-form-group">
										<label style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '4px' }}>Pin Code</label>
										<input
											type="text"
											className="b2b-input"
											value={sameAsBilling ? billingPincode : shippingPincode}
											onChange={(e) => setShippingPincode(e.target.value)}
											placeholder="6-digit PIN"
											maxLength={6}
											disabled={sameAsBilling}
											style={{ opacity: sameAsBilling ? 0.6 : 1 }}
										/>
									</div>
								</div>
							</div>
						</div>

						<div style={{ marginTop: '20px', borderTop: '1px solid var(--border-color)', paddingTop: '16px' }}>
							<div className="sync-form-group">
								<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Notes & Other Details</label>
								<textarea
									className="b2b-input"
									rows={3}
									style={{ resize: 'vertical', minHeight: '80px', fontFamily: 'inherit' }}
									value={editingCustomer.notes || ''}
									onChange={(e) => setEditingCustomer({ ...editingCustomer, notes: e.target.value })}
									placeholder="Add internal remarks, secondary contacts, payment terms, or shipping guidelines for this client..."
								/>
							</div>
						</div>

						<div className="modal-actions" style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '24px', borderTop: '1px solid var(--border-color)', paddingTop: '20px' }}>
							<button onClick={() => {
								setShowCustomerModal(false);
								setEditingCustomer(null);
							}} className="b2b-btn b2b-btn-secondary" style={{ minWidth: '120px' }}>Cancel</button>
							<button onClick={() => handleSaveCustomer(editingCustomer)} className="b2b-btn b2b-btn-primary" style={{ minWidth: '140px' }}>Save Client</button>
						</div>
					</div>
				</div>,
				document.body
			)}

			{/* PAYMENT ENTRY MODAL */}
			{showPaymentModal && paymentInvoice && createPortal(
				<div className="modal-overlay" onClick={() => {
					setShowPaymentModal(false);
					setPaymentInvoice(null);
				}} style={{ alignItems: 'flex-start', overflowY: 'auto', padding: '2rem 1.5rem' }}>
					<div className="premium-modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '420px', width: '100%', margin: '0 auto' }}>
						<h2>Record Payment</h2>
						<p>Record a client transaction for this invoice below.</p>

						<div style={{ fontSize: '0.85rem', background: 'var(--bg-input)', padding: '14px', borderRadius: '10px', marginBottom: '20px', lineHeight: 1.6, border: '1px solid var(--border-color)' }}>
							<strong>Invoice:</strong> {paymentInvoice.invoice_number}<br />
							<strong>Total Price:</strong> ₹{paymentInvoice.total_price.toFixed(2)}<br />
							<strong>Current Balance:</strong> ₹{paymentInvoice.balance_amount.toFixed(2)}
						</div>

						<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
							<div className="sync-form-group">
								<label>Paid Amount (₹)</label>
								<input
									type="number"
									className="b2b-input"
									value={paymentAmount}
									onChange={(e) => setPaymentAmount(Number(e.target.value))}
									max={paymentInvoice.total_price}
								/>
							</div>

							<div className="sync-form-group">
								<label>Payment Method</label>
								<select
									value={paymentMethod}
									onChange={(e) => setPaymentMethod(e.target.value)}
									className="b2b-input"
								>
									<option value="Bank Transfer">Bank Transfer</option>
									<option value="Cash">Cash</option>
									<option value="UPI">UPI</option>
									<option value="Cheque">Cheque</option>
								</select>
							</div>
						</div>

						<div className="modal-actions" style={{ marginTop: '28px' }}>
							<button onClick={() => {
								setShowPaymentModal(false);
								setPaymentInvoice(null);
							}} className="btn-secondary">Cancel</button>
							<button onClick={handleSavePayment} className="btn-primary">Record Payment</button>
						</div>
					</div>
				</div>,
				document.body
			)}

			{/* NEW PAYMENT TERM MODAL */}
			{showPaymentTermModal && createPortal(
				<div className="modal-overlay" onClick={() => {
					setShowPaymentTermModal(false);
					setNewTermName('');
					setNewTermDays(0);
				}} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', overflowY: 'auto', padding: '2rem 1.5rem', zIndex: 3100 }}>
					<div className="premium-modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '450px', width: '100%', margin: '0 auto', border: '1px solid var(--border-color)', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.15)', borderRadius: '24px', padding: '2.5rem', background: 'var(--surface-color)', position: 'relative' }}>

						{/* Modal Header */}
						<div style={{ display: 'flex', alignItems: 'center', gap: '16px', marginBottom: '2rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1.25rem' }}>
							<div style={{
								width: '48px',
								height: '48px',
								borderRadius: '12px',
								background: 'linear-gradient(135deg, var(--accent-color), #0ea5e9)',
								display: 'flex',
								alignItems: 'center',
								justifyContent: 'center',
								color: 'white',
								boxShadow: '0 8px 16px rgba(14, 165, 233, 0.15)',
								flexShrink: 0
							}}>
								<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
									<line x1="16" y1="2" x2="16" y2="6" />
									<line x1="8" y1="2" x2="8" y2="6" />
									<line x1="3" y1="10" x2="21" y2="10" />
								</svg>
							</div>
							<div>
								<h2 style={{ margin: 0, fontSize: '1.45rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>New Payment Term</h2>
								<p style={{ margin: '3px 0 0 0', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>Configure payment timelines for B2B billing.</p>
							</div>
						</div>

						{/* Yellow Banner Notification */}
						<div style={{ background: 'var(--accent-subtle)', padding: '14px', borderRadius: '12px', border: '1px solid rgba(14, 165, 233, 0.15)', color: 'var(--text-secondary)', fontSize: '0.85rem', marginBottom: '20px', display: 'flex', alignItems: 'center', gap: '8px', lineHeight: 1.4 }}>
							<span>💡</span>
							<span>The Payment Terms can now be configured and managed from <strong>Settings &rarr; Setup & Configuration &rarr; Payment Terms</strong>.</span>
						</div>

						<div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
							<div className="sync-form-group">
								<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Term Name*</label>
								<input
									type="text"
									className="b2b-input"
									value={newTermName}
									onChange={(e) => setNewTermName(e.target.value)}
									placeholder="e.g. Net 45"
								/>
							</div>

							<div className="sync-form-group">
								<label style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: '6px' }}>Due After*</label>
								<div style={{ display: 'flex', alignItems: 'center' }}>
									<input
										type="number"
										className="b2b-input days-input-left"
										value={newTermDays}
										onChange={(e) => setNewTermDays(Number(e.target.value))}
										min="0"
									/>
									<span className="days-label-right">Days</span>
								</div>
							</div>
						</div>

						<div className="modal-actions" style={{ marginTop: '28px' }}>
							<button onClick={() => {
								setShowPaymentTermModal(false);
								setNewTermName('');
								setNewTermDays(0);
							}} className="b2b-btn b2b-btn-secondary" style={{ flex: 1 }}>Cancel</button>
							<button onClick={handleSavePaymentTerm} className="b2b-btn b2b-btn-primary" style={{ flex: 1 }}>Save</button>
						</div>
					</div>
				</div>,
				document.body
			)}

			{/* CUSTOMER LEDGER STATEMENT MODAL */}
			{showLedgerModal && ledgerCustomer && createPortal(
				<div className="modal-overlay" onClick={() => {
					setShowLedgerModal(false);
					setLedgerCustomer(null);
					setLedgerData(null);
				}} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', overflowY: 'auto', padding: '2rem 1.5rem', zIndex: 3200 }}>
					<div className="premium-modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '850px', width: '100%', margin: '0 auto', border: '1px solid var(--border-color)', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.15)', borderRadius: '24px', padding: '2.5rem', background: 'var(--surface-color)', position: 'relative' }}>
						
						{/* Modal Header */}
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '2rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '1.25rem' }}>
							<div>
								<h2 style={{ margin: 0, fontSize: '1.45rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>Customer Ledger Statement</h2>
								<p style={{ margin: '3px 0 0 0', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
									Chronological account history for <strong>{ledgerCustomer.legal_name}</strong> ({ledgerCustomer.gstin})
								</p>
							</div>
							<button onClick={() => {
								setShowLedgerModal(false);
								setLedgerCustomer(null);
								setLedgerData(null);
							}} className="b2b-btn b2b-btn-secondary" style={{ padding: '6px 12px' }}>✕ Close</button>
						</div>

						{ledgerData ? (
							<div>
								{/* Balance Overview Cards */}
								<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px', marginBottom: '24px' }}>
									<div style={{ background: 'var(--bg-hover)', padding: '16px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
										<div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Opening Balance</div>
										<div style={{ fontSize: '1.4rem', fontWeight: 800, color: 'var(--text-primary)', marginTop: '4px' }}>₹{ledgerData.opening_balance.toFixed(2)}</div>
									</div>
									<div style={{ background: 'linear-gradient(135deg, rgba(99,102,241,0.05) 0%, rgba(16,185,129,0.05) 100%)', padding: '16px', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
										<div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-tertiary)', textTransform: 'uppercase' }}>Closing Outstanding Balance</div>
										<div style={{ fontSize: '1.4rem', fontWeight: 800, color: 'var(--accent-color)', marginTop: '4px' }}>₹{ledgerData.closing_balance.toFixed(2)}</div>
									</div>
								</div>

								{/* Ledger Table */}
								<div style={{ maxHeight: '350px', overflowY: 'auto', borderRadius: '12px', border: '1px solid var(--border-color)' }}>
									<table className="b2b-table" style={{ width: '100%' }}>
										<thead>
											<tr>
												<th>Date</th>
												<th>Type</th>
												<th>Reference</th>
												<th>Debit (Dr)</th>
												<th>Credit (Cr)</th>
												<th>Running Balance</th>
											</tr>
										</thead>
										<tbody>
											{ledgerData.transactions && ledgerData.transactions.length === 0 ? (
												<tr>
													<td colSpan={6} style={{ textAlign: 'center', padding: '30px', color: 'var(--text-tertiary)' }}>No financial transactions found.</td>
												</tr>
											) : (
												ledgerData.transactions && ledgerData.transactions.map((tx: any, idx: number) => (
													<tr key={idx}>
														<td>{tx.date ? tx.date.split('T')[0] : ''}</td>
														<td>
															<span style={{
																padding: '4px 8px',
																borderRadius: '8px',
																fontSize: '11px',
																fontWeight: 600,
																background: tx.type === 'INVOICE' || tx.type === 'DEBIT_NOTE' ? 'rgba(99,102,241,0.1)' : 'rgba(16,185,129,0.1)',
																color: tx.type === 'INVOICE' || tx.type === 'DEBIT_NOTE' ? 'rgb(99,102,241)' : 'rgb(16,185,129)'
															}}>
																{tx.type}
															</span>
														</td>
														<td>{tx.reference}</td>
														<td style={{ color: tx.debit > 0 ? 'var(--text-primary)' : 'var(--text-tertiary)' }}>{tx.debit > 0 ? `₹${tx.debit.toFixed(2)}` : '—'}</td>
														<td style={{ color: tx.credit > 0 ? 'var(--status-active)' : 'var(--text-tertiary)' }}>{tx.credit > 0 ? `₹${tx.credit.toFixed(2)}` : '—'}</td>
														<td style={{ fontWeight: 'bold' }}>₹{tx.running_balance.toFixed(2)}</td>
													</tr>
												))
											)}
										</tbody>
									</table>
								</div>
							</div>
						) : (
							<div style={{ textAlign: 'center', padding: '40px' }}>Loading ledger statement...</div>
						)}
					</div>
				</div>,
				document.body
			)}

			{/* CUSTOM CONFIRM MODAL */}
			{confirmConfig.show && createPortal(
				<div className="modal-overlay" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 9999, backgroundColor: 'rgba(15, 23, 42, 0.45)', backdropFilter: 'blur(8px)' }}>
					<div className="premium-modal" style={{ maxWidth: '440px', width: '90%', padding: '28px', borderRadius: '24px', border: '1px solid var(--border-color)', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)', background: 'var(--surface-color)', animation: 'modalFadeIn 0.2s ease-out', position: 'relative' }}>
						<div style={{ display: 'flex', alignItems: 'flex-start', gap: '16px', marginBottom: '20px' }}>
							<div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: '48px', height: '48px', borderRadius: '16px', background: 'rgba(239, 68, 68, 0.08)', color: '#ef4444', flexShrink: 0 }}>
								<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
									<circle cx="12" cy="12" r="10"></circle>
									<line x1="12" y1="8" x2="12" y2="12"></line>
									<line x1="12" y1="16" x2="12.01" y2="16"></line>
								</svg>
							</div>
							<div style={{ flex: 1 }}>
								<h3 style={{ margin: 0, fontSize: '1.25rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>Action Required</h3>
								<p style={{ margin: '8px 0 0 0', fontSize: '0.925rem', color: 'var(--text-secondary)', lineHeight: '1.5', fontWeight: 500 }}>
									{confirmConfig.message}
								</p>
							</div>
						</div>
						<div className="modal-actions" style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end', marginTop: '24px', borderTop: '1px solid var(--border-color)', paddingTop: '20px' }}>
							<button 
								onClick={() => {
									confirmConfig.resolve?.(false);
									setConfirmConfig({ show: false, message: '' });
								}} 
								className="b2b-btn b2b-btn-secondary" 
								style={{ flex: 1, padding: '10px 16px', borderRadius: '12px', fontSize: '0.85rem', fontWeight: 600 }}
							>
								Cancel
							</button>
							<button 
								onClick={() => {
									confirmConfig.resolve?.(true);
									setConfirmConfig({ show: false, message: '' });
								}} 
								className="b2b-btn b2b-btn-danger" 
								style={{ flex: 1, padding: '10px 16px', borderRadius: '12px', fontSize: '0.85rem', fontWeight: 600, background: 'var(--status-danger)', color: '#ffffff', borderColor: 'var(--status-danger)' }}
							>
								Confirm
							</button>
						</div>
					</div>
				</div>,
				document.body
			)}

			{/* CUSTOM ALERT MODAL */}
			{alertConfig.show && createPortal(
				<div className="modal-overlay" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 10000, backgroundColor: 'rgba(15, 23, 42, 0.45)', backdropFilter: 'blur(8px)' }}>
					<div className="premium-modal" style={{ maxWidth: '440px', width: '90%', padding: '28px', borderRadius: '24px', border: '1px solid var(--border-color)', boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)', background: 'var(--surface-color)', animation: 'modalFadeIn 0.2s ease-out', position: 'relative' }}>
						<div style={{ display: 'flex', alignItems: 'flex-start', gap: '16px', marginBottom: '20px' }}>
							{/success/i.test(alertConfig.message) ? (
								// Success Icon
								<div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: '48px', height: '48px', borderRadius: '16px', background: 'rgba(16, 185, 129, 0.08)', color: '#10b981', flexShrink: 0 }}>
									<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
										<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
										<polyline points="22 4 12 14.01 9 11.01"></polyline>
									</svg>
								</div>
							) : (
								// Warning/Info Icon
								<div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: '48px', height: '48px', borderRadius: '16px', background: 'rgba(99, 102, 241, 0.08)', color: 'var(--accent-color)', flexShrink: 0 }}>
									<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
										<circle cx="12" cy="12" r="10"></circle>
										<line x1="12" y1="16" x2="12" y2="12"></line>
										<line x1="12" y1="8" x2="12.01" y2="8"></line>
									</svg>
								</div>
							)}
							<div style={{ flex: 1 }}>
								<h3 style={{ margin: 0, fontSize: '1.25rem', fontWeight: 800, color: 'var(--text-primary)', letterSpacing: '-0.02em' }}>
									{/success/i.test(alertConfig.message) ? 'Success' : 'Notification'}
								</h3>
								<p style={{ margin: '8px 0 0 0', fontSize: '0.925rem', color: 'var(--text-secondary)', lineHeight: '1.5', fontWeight: 500 }}>
									{alertConfig.message}
								</p>
							</div>
						</div>
						<div className="modal-actions" style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end', marginTop: '24px', borderTop: '1px solid var(--border-color)', paddingTop: '20px' }}>
							<button 
								onClick={() => setAlertConfig({ show: false, message: '' })} 
								className="b2b-btn b2b-btn-primary" 
								style={{ flex: 1, padding: '10px 16px', borderRadius: '12px', fontSize: '0.85rem', fontWeight: 600, background: /success/i.test(alertConfig.message) ? '#10b981' : 'var(--accent-color)', color: '#ffffff', borderColor: /success/i.test(alertConfig.message) ? '#10b981' : 'var(--accent-color)' }}
							>
								OK
							</button>
						</div>
					</div>
				</div>,
				document.body
			)}
		</>
	);
}
