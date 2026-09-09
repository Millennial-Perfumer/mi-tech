import { useCallback, useEffect, useState } from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  Check,
  CircleAlert,
  Clock3,
  ExternalLink,
  RefreshCw,
  Search,
  Send,
  ShoppingBag,
  Trash2,
  X,
} from "lucide-react";
import { API_BASE } from "../../lib/api";
import { usePeriodFilter } from "../../lib/usePeriodFilter";

type AbandonedCartsPageProps = { token: string; onUnauthorized: () => void };

type LineItem = {
  title?: string;
  variant_title?: string;
  sku?: string;
  quantity?: number | string;
  price?: number | string;
};

type Checkout = {
  id: number;
  store_id?: string;
  checkout_id: string;
  checkout_token?: string;
  cart_token?: string;
  email: string;
  phone: string;
  customer_name: string;
  checkout_url: string;
  line_items: unknown;
  total_price: number;
  currency: string;
  completed: boolean;
  completed_at?: string;
  order_id?: string;
  recovery_status: string;
  recovery_attempts: number;
  recovery_message_sent_at?: string;
  last_error?: string;
  marketing_consent: boolean;
  sms_consent?: boolean;
  city?: string;
  province?: string;
  country?: string;
  zip?: string;
  abandoned_at: string;
  created_at?: string;
  updated_at?: string;
};

type WhatsappStats = {
  sent: number;
  delivered: number;
  read: number;
  clicked: number;
  failed: number;
};

type RevenueTimelineItem = {
  date: string;
  abandonedAmount: number;
  recoveredAmount: number;
  abandonedCount?: number;
  recoveredCount?: number;
};

type StatusBreakdownItem = {
  status: string;
  count: number;
  amount: number;
};

type Analytics = {
  totalAbandonedRevenue: number;
  recoveredRevenue: number;
  pendingRevenue: number;
  abandonedCartCount: number;
  recoveredCartCount: number;
  recoveryRate: number;
  cartsCreatedCount?: number;
  addCartToCheckoutRate?: number;
  addCartToOrderRate?: number;
  averageCartValue?: number;
  contactableCartCount?: number;
  contactabilityRate?: number;
  averageRecoveryTimeMinutes?: number;
  averageAttemptsToRecovery?: number;
  marketingConsentCount?: number;
  smsConsentCount?: number;
  whatsappStats?: WhatsappStats;
  revenueTimeline?: RevenueTimelineItem[];
  statusBreakdown?: StatusBreakdownItem[];
  topLostCarts?: {
    id?: number;
    customer_name: string;
    phone: string;
    total_price: number;
    currency: string;
    abandoned_at: string;
    recovery_status: string;
    recovery_attempts?: number;
    attempts?: number;
  }[];
};

const statusOptions = [
  "PENDING",
  "PROCESSING",
  "SENT",
  "FAILED",
  "CANCELLED",
  "RECOVERED",
];
const statusColors: Record<string, string> = {
  Recovered: "#10b981",
  "Message Sent": "#3b82f6",
  Pending: "#f59e0b",
  Failed: "#ef4444",
  Expired: "#6b7280",
};

function numberValue(value: number | string | undefined) {
  const parsed = typeof value === "number" ? value : Number(value || 0);
  return Number.isFinite(parsed) ? parsed : 0;
}

function money(value: number | string | undefined, currency = "INR") {
  const code = /^[A-Z]{3}$/.test(currency) ? currency : "INR";
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: code,
  }).format(numberValue(value));
}

function percentage(value: number, denominator: number) {
  return denominator > 0 ? (value / denominator) * 100 : 0;
}

function dateTime(value: string | undefined) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return date.toLocaleString("en-IN", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

function dateLabel(value: string) {
  const date = new Date(`${value}T12:00:00`);
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleDateString("en-IN", { day: "numeric", month: "short" });
}

function parseLineItems(raw: unknown): LineItem[] {
  if (!raw) return [];
  try {
    const parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
    return Array.isArray(parsed) ? (parsed as LineItem[]) : [];
  } catch {
    return [];
  }
}

function itemQuantity(item: LineItem) {
  return numberValue(item.quantity || 1);
}

function itemLabel(item: LineItem) {
  const title = item.title || "Product";
  return item.variant_title ? `${title} — ${item.variant_title}` : title;
}

function lineItemSummary(raw: unknown) {
  const items = parseLineItems(raw);
  const quantity = items.reduce((total, item) => total + itemQuantity(item), 0);
  const summary = items
    .map((item) => `${item.title || "Product"} (×${itemQuantity(item)})`)
    .join(", ");
  return { quantity, summary };
}

function formatPhone(value: string | undefined) {
  if (!value) return "—";
  const digits = value.replace(/\D/g, "");
  if (!digits) return value;
  const normalized = digits.length === 10 ? `91${digits}` : digits;
  return `+${normalized.slice(0, 2)} ${normalized.slice(2, 7)} ${normalized.slice(7)}`;
}

function formatDuration(minutes: number | undefined) {
  const value = numberValue(minutes);
  if (value <= 0) return "Not enough recovered data";
  if (value < 60) return `${Math.round(value)} min`;
  const hours = Math.floor(value / 60);
  const remainder = Math.round(value % 60);
  return remainder ? `${hours}h ${remainder}m` : `${hours}h`;
}

export function AbandonedCartsPage({
  token,
  onUnauthorized,
}: AbandonedCartsPageProps) {
  const { startDate, endDate } = usePeriodFilter();
  const [activeTab, setActiveTab] = useState<"list" | "analytics">("list");
  const [checkouts, setCheckouts] = useState<Checkout[]>([]);
  const [analytics, setAnalytics] = useState<Analytics | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");
  const [selected, setSelected] = useState<Checkout | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAnalyticsLoading, setIsAnalyticsLoading] = useState(false);
  const [isWorking, setIsWorking] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const request = useCallback(
    async (path: string, options: RequestInit = {}) => {
      const response = await fetch(`${API_BASE}${path}`, {
        ...options,
        headers: {
          Authorization: `Bearer ${token}`,
          ...(options.headers || {}),
        },
      });
      if (response.status === 401) {
        onUnauthorized();
        throw new Error("Your session has expired. Please sign in again.");
      }
      if (!response.ok)
        throw new Error(
          `Abandoned carts request failed with status ${response.status}`,
        );
      return response;
    },
    [onUnauthorized, token],
  );

  const loadList = useCallback(async () => {
    setIsLoading(true);
    setError("");
    const query = new URLSearchParams({
      page: String(page),
      limit: "15",
      search,
      start_date: startDate,
      end_date: endDate,
    });
    if (status) query.set("status", status);
    try {
      const response = await request(
        `/api/abandoned-checkouts?${query.toString()}`,
      );
      const data = (await response.json()) as {
        checkouts?: Checkout[];
        total_count?: number;
      };
      setCheckouts(data.checkouts || []);
      setTotal(data.total_count || 0);
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Unable to load abandoned carts",
      );
    } finally {
      setIsLoading(false);
    }
  }, [endDate, page, request, search, startDate, status]);

  const loadAnalytics = useCallback(async () => {
    setIsAnalyticsLoading(true);
    setError("");
    try {
      const query = new URLSearchParams({
        start_date: startDate,
        end_date: endDate,
      });
      const response = await request(
        `/api/abandoned-checkouts/analytics?${query.toString()}`,
      );
      const data = (await response.json()) as { analytics?: Analytics };
      setAnalytics(data.analytics || null);
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Unable to load recovery analytics",
      );
    } finally {
      setIsAnalyticsLoading(false);
    }
  }, [endDate, request, startDate]);

  useEffect(() => {
    if (activeTab === "list") void loadList();
  }, [activeTab, loadList]);

  useEffect(() => {
    if (activeTab === "analytics") void loadAnalytics();
  }, [activeTab, loadAnalytics]);

  useEffect(() => {
    setPage(1);
  }, [endDate, search, startDate, status]);

  const totalPages = Math.max(Math.ceil(total / 15), 1);

  const dispatchRecovery = async (id: number) => {
    setIsWorking(true);
    setError("");
    try {
      await request(`/api/abandoned-checkouts/recover?id=${id}`, {
        method: "POST",
      });
      const sentAt = new Date().toISOString();
      setNotice("Recovery message dispatched");
      setSelected((current) =>
        current?.id === id
          ? {
              ...current,
              recovery_status: "SENT",
              recovery_attempts: current.recovery_attempts + 1,
              recovery_message_sent_at: sentAt,
            }
          : current,
      );
      if (activeTab === "list") await loadList();
      else await loadAnalytics();
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Unable to dispatch recovery",
      );
    } finally {
      setIsWorking(false);
    }
  };

  const recover = async (checkout: Checkout) => dispatchRecovery(checkout.id);

  const updateStatus = async (checkout: Checkout, nextStatus: string) => {
    setIsWorking(true);
    setError("");
    const completed = nextStatus === "RECOVERED";
    const recoveryStatus = completed ? "SENT" : nextStatus;
    try {
      await request("/api/abandoned-checkouts/status", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: checkout.id,
          recovery_status: recoveryStatus,
          completed,
        }),
      });
      setNotice("Checkout status updated");
      setSelected((current) =>
        current?.id === checkout.id
          ? {
              ...current,
              recovery_status: recoveryStatus,
              completed,
              completed_at: completed ? new Date().toISOString() : undefined,
              order_id: completed ? current.order_id : undefined,
            }
          : current,
      );
      await loadList();
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Unable to update checkout status",
      );
    } finally {
      setIsWorking(false);
    }
  };

  const deleteCheckout = async (checkout: Checkout) => {
    if (!window.confirm(`Delete checkout ${checkout.checkout_id}?`)) return;
    setIsWorking(true);
    setError("");
    try {
      await request(`/api/abandoned-checkouts?id=${checkout.id}`, {
        method: "DELETE",
      });
      setSelected(null);
      setNotice("Checkout record deleted");
      await loadList();
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Unable to delete checkout",
      );
    } finally {
      setIsWorking(false);
    }
  };

  const refresh = activeTab === "list" ? loadList : loadAnalytics;

  return (
    <section
      className="workspace-page abandoned-page"
      aria-labelledby="abandoned-heading"
    >
      <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Engagement / Recovery</p>
          <h2 id="abandoned-heading">Abandoned carts</h2>
          <p>
            Review abandoned checkouts, cart drop-off ratios, and recovery
            performance.
          </p>
        </div>
        <button
          className="secondary-button"
          type="button"
          onClick={() => void refresh()}
          disabled={isLoading || isAnalyticsLoading}
        >
          <RefreshCw
            size={15}
            className={isLoading || isAnalyticsLoading ? "spin" : undefined}
            aria-hidden="true"
          />
          Refresh
        </button>
      </header>

      {error && (
        <div className="dashboard-error" role="alert">
          <CircleAlert size={18} aria-hidden="true" />
          <span>{error}</span>
          <button type="button" onClick={() => void refresh()}>
            Try again
          </button>
        </div>
      )}
      {notice && (
        <div className="inventory-notice" role="status">
          {notice}
        </div>
      )}

      <div
        className="abandoned-tabs"
        role="tablist"
        aria-label="Abandoned checkout views"
      >
        <button
          className={activeTab === "list" ? "abandoned-tab-active" : ""}
          type="button"
          role="tab"
          aria-selected={activeTab === "list"}
          onClick={() => setActiveTab("list")}
        >
          <ShoppingBag size={15} aria-hidden="true" /> Checkout list
        </button>
        <button
          className={activeTab === "analytics" ? "abandoned-tab-active" : ""}
          type="button"
          role="tab"
          aria-selected={activeTab === "analytics"}
          onClick={() => setActiveTab("analytics")}
        >
          <span aria-hidden="true">↗</span> Recovery analytics
        </button>
      </div>

      {activeTab === "list" ? (
        <>
          <div className="abandoned-toolbar">
            <label className="orders-search">
              <Search size={16} aria-hidden="true" />
              <span className="sr-only">Search abandoned checkouts</span>
              <input
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Search customer, phone, email, or checkout"
              />
            </label>
            <label className="compact-select">
              <span className="sr-only">Recovery status</span>
              <select
                value={status}
                onChange={(event) => setStatus(event.target.value)}
              >
                <option value="">All statuses</option>
                {statusOptions.map((option) => (
                  <option key={option} value={option}>
                    {option.replace("_", " ")}
                  </option>
                ))}
              </select>
            </label>
            <span className="filter-count">
              {total.toLocaleString("en-IN")} checkouts
            </span>
          </div>

          <div className="abandoned-list">
            {isLoading ? (
              <div className="empty-panel">
                <ShoppingBag size={20} aria-hidden="true" />
                <p>Loading abandoned checkouts…</p>
              </div>
            ) : checkouts.length === 0 ? (
              <div className="empty-panel">
                <ShoppingBag size={20} aria-hidden="true" />
                <div>
                  <h2>No checkouts found</h2>
                  <p>Try another status or search term.</p>
                </div>
              </div>
            ) : (
              checkouts.map((checkout) => {
                const items = lineItemSummary(checkout.line_items);
                const contact = [
                  checkout.email,
                  checkout.phone ? formatPhone(checkout.phone) : "",
                ]
                  .filter(Boolean)
                  .join(" · ");
                const location = [
                  checkout.city,
                  checkout.province,
                  checkout.country,
                  checkout.zip,
                ]
                  .filter(Boolean)
                  .join(", ");
                return (
                  <article className="abandoned-card" key={checkout.id}>
                    <div className="abandoned-card-main">
                      <div className="abandoned-card-heading">
                        <div>
                          <span className="checkout-id">
                            {checkout.checkout_id || "Checkout"}
                          </span>
                          <h3>
                            {checkout.customer_name || "Anonymous customer"}
                          </h3>
                        </div>
                        <strong>
                          {money(checkout.total_price, checkout.currency)}
                        </strong>
                      </div>
                      <p>
                        {contact || "No contact details"} · Abandoned{" "}
                        {dateTime(checkout.abandoned_at)}
                        {location ? ` · ${location}` : ""}
                      </p>
                      <div className="abandoned-card-cart-summary">
                        <strong>
                          {items.quantity}{" "}
                          {items.quantity === 1 ? "item" : "items"}
                        </strong>
                        <span title={items.summary}>
                          {items.summary || "No line item details"}
                        </span>
                      </div>
                      <div className="abandoned-card-foot">
                        <span
                          className={`status-pill status-pill-${checkout.completed ? "success" : checkout.recovery_status === "FAILED" ? "danger" : "neutral"}`}
                        >
                          {checkout.completed
                            ? "Recovered"
                            : checkout.recovery_status.replace("_", " ")}
                        </span>
                        <span>
                          <Clock3 size={13} aria-hidden="true" />{" "}
                          {checkout.recovery_attempts} recovery attempts
                        </span>
                        {checkout.marketing_consent && (
                          <span>Marketing consent</span>
                        )}
                        {checkout.sms_consent && <span>SMS consent</span>}
                      </div>
                    </div>
                    <div className="abandoned-card-actions">
                      <label className="compact-select">
                        <span className="sr-only">
                          Status for {checkout.checkout_id}
                        </span>
                        <select
                          value={
                            checkout.completed
                              ? "RECOVERED"
                              : checkout.recovery_status
                          }
                          disabled={isWorking}
                          onChange={(event) =>
                            void updateStatus(checkout, event.target.value)
                          }
                        >
                          {statusOptions.map((option) => (
                            <option key={option} value={option}>
                              {option.replace("_", " ")}
                            </option>
                          ))}
                        </select>
                      </label>
                      <button
                        className="secondary-button"
                        type="button"
                        onClick={() => setSelected(checkout)}
                      >
                        Details
                      </button>
                      {!checkout.completed && (
                        <button
                          className="primary-button"
                          type="button"
                          onClick={() => void recover(checkout)}
                          disabled={isWorking || !checkout.phone}
                        >
                          <Send size={14} aria-hidden="true" /> Recover
                        </button>
                      )}
                    </div>
                  </article>
                );
              })
            )}
          </div>

          <div className="orders-pagination">
            <span>
              Showing {checkouts.length ? (page - 1) * 15 + 1 : 0}–
              {Math.min(page * 15, total)} of {total}
            </span>
            <div>
              <button
                type="button"
                aria-label="Previous checkout page"
                disabled={page <= 1 || isLoading}
                onClick={() => setPage((current) => current - 1)}
              >
                ‹
              </button>
              <span className="pagination-current-page">
                Page {page} of {totalPages}
              </span>
              <button
                type="button"
                aria-label="Next checkout page"
                disabled={page >= totalPages || isLoading}
                onClick={() => setPage((current) => current + 1)}
              >
                ›
              </button>
            </div>
          </div>
        </>
      ) : (
        <AnalyticsPanel
          analytics={analytics}
          isLoading={isAnalyticsLoading}
          onRecover={(id, phone) => {
            const checkout = checkouts.find((item) => item.id === id);
            if (checkout) void recover(checkout);
            else if (phone) void dispatchRecovery(id);
          }}
          sending={isWorking}
        />
      )}

      {selected && (
        <CheckoutDetailsModal
          checkout={selected}
          isWorking={isWorking}
          onClose={() => setSelected(null)}
          onDelete={() => void deleteCheckout(selected)}
          onRecover={() => void recover(selected)}
          onStatusChange={(nextStatus) =>
            void updateStatus(selected, nextStatus)
          }
        />
      )}
    </section>
  );
}

function CheckoutDetailsModal({
  checkout,
  isWorking,
  onClose,
  onDelete,
  onRecover,
  onStatusChange,
}: {
  checkout: Checkout;
  isWorking: boolean;
  onClose: () => void;
  onDelete: () => void;
  onRecover: () => void;
  onStatusChange: (status: string) => void;
}) {
  const items = parseLineItems(checkout.line_items);
  const location = [
    checkout.city,
    checkout.province,
    checkout.country,
    checkout.zip,
  ]
    .filter(Boolean)
    .join(", ");
  const currentStatus = checkout.completed
    ? "RECOVERED"
    : checkout.recovery_status;

  return (
    <div
      className="modal-scrim"
      role="presentation"
      onMouseDown={(event) => event.target === event.currentTarget && onClose()}
    >
      <div
        className="modal-card abandoned-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="checkout-detail-heading"
      >
        <div className="modal-heading">
          <div>
            <p className="eyebrow">Checkout details</p>
            <h2 id="checkout-detail-heading">
              {checkout.customer_name || "Anonymous customer"}
            </h2>
            <span className="checkout-id">
              {checkout.checkout_id || "No checkout ID"}
            </span>
          </div>
          <button
            className="icon-button"
            type="button"
            aria-label="Close checkout details"
            onClick={onClose}
          >
            <X size={19} aria-hidden="true" />
          </button>
        </div>

        <div className="checkout-detail-grid">
          <span>
            Checkout ID<strong>{checkout.checkout_id || "—"}</strong>
          </span>
          <span>
            Contact
            <strong>{checkout.email || formatPhone(checkout.phone)}</strong>
          </span>
          <span>
            Phone<strong>{formatPhone(checkout.phone)}</strong>
          </span>
          <span>
            Location<strong>{location || "—"}</strong>
          </span>
          <span>
            Cart token<strong>{checkout.cart_token || "—"}</strong>
          </span>
          <span>
            Abandoned<strong>{dateTime(checkout.abandoned_at)}</strong>
          </span>
          <span>
            Created<strong>{dateTime(checkout.created_at)}</strong>
          </span>
          <span>
            Recovery attempts<strong>{checkout.recovery_attempts}</strong>
          </span>
          <span>
            Message sent
            <strong>{dateTime(checkout.recovery_message_sent_at)}</strong>
          </span>
          <span>
            Recovered<strong>{dateTime(checkout.completed_at)}</strong>
          </span>
          <span>
            Order ID
            <strong>{checkout.order_id ? `#${checkout.order_id}` : "—"}</strong>
          </span>
          <span>
            Consent
            <strong>
              {checkout.marketing_consent ? "Marketing" : "No marketing"}
              {checkout.sms_consent ? " · SMS" : ""}
            </strong>
          </span>
        </div>

        <div className="abandoned-detail-status-row">
          <label className="compact-select">
            <span>Status</span>
            <select
              value={currentStatus}
              disabled={isWorking}
              onChange={(event) => onStatusChange(event.target.value)}
            >
              {statusOptions.map((option) => (
                <option key={option} value={option}>
                  {option.replace("_", " ")}
                </option>
              ))}
            </select>
          </label>
          {checkout.checkout_url && (
            <a
              className="secondary-button"
              href={checkout.checkout_url}
              target="_blank"
              rel="noreferrer"
            >
              <ExternalLink size={14} aria-hidden="true" /> Open checkout link
            </a>
          )}
        </div>

        <h3 className="detail-section-heading">Line items ({items.length})</h3>
        <div className="checkout-line-items">
          {items.length ? (
            items.map((item, index) => (
              <div key={`${item.title || "product"}-${index}`}>
                <span>
                  <strong>{itemLabel(item)}</strong>
                  <small>
                    {item.sku ? `SKU: ${item.sku} · ` : ""}×{" "}
                    {itemQuantity(item)}
                  </small>
                </span>
                <strong>
                  {money(
                    numberValue(item.price) * itemQuantity(item),
                    checkout.currency,
                  )}
                </strong>
              </div>
            ))
          ) : (
            <p>No line item details available.</p>
          )}
        </div>

        {checkout.last_error && (
          <div className="dashboard-error">
            <CircleAlert size={16} aria-hidden="true" />
            <span>
              <strong>Last error:</strong> {checkout.last_error}
            </span>
          </div>
        )}

        <div className="modal-actions">
          <button
            className="secondary-button"
            type="button"
            onClick={onDelete}
            disabled={isWorking}
          >
            <Trash2 size={14} aria-hidden="true" /> Delete record
          </button>
          <button className="secondary-button" type="button" onClick={onClose}>
            Close
          </button>
          {!checkout.completed && (
            <button
              className="primary-button"
              type="button"
              onClick={onRecover}
              disabled={isWorking || !checkout.phone}
            >
              <Send size={14} aria-hidden="true" /> Recover via WhatsApp
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

function AnalyticsPanel({
  analytics,
  isLoading,
  onRecover,
  sending,
}: {
  analytics: Analytics | null;
  isLoading: boolean;
  onRecover: (id: number, phone: string) => void;
  sending: boolean;
}) {
  if (isLoading) {
    return (
      <div className="abandoned-analytics-loading">
        <div className="report-metrics-grid">
          {[1, 2, 3, 4, 5, 6, 7, 8].map((item) => (
            <div className="report-metric-card skeleton-loader" key={item} />
          ))}
        </div>
        <div className="analytics-grid">
          <div className="reports-table-card abandoned-chart-card skeleton-loader" />
          <div className="reports-table-card abandoned-chart-card skeleton-loader" />
        </div>
      </div>
    );
  }

  if (!analytics) {
    return (
      <div className="empty-panel">
        <ShoppingBag size={20} aria-hidden="true" />
        <div>
          <h2>No analytics data</h2>
          <p>No cart or checkout records were found for this period.</p>
        </div>
      </div>
    );
  }

  const whatsapp: WhatsappStats = {
    sent: 0,
    delivered: 0,
    read: 0,
    clicked: 0,
    failed: 0,
    ...analytics.whatsappStats,
  };
  const abandonedCount = numberValue(analytics.abandonedCartCount);
  const recoveredCount = numberValue(analytics.recoveredCartCount);
  const cartsCreated = Math.max(
    numberValue(analytics.cartsCreatedCount),
    abandonedCount,
  );
  const cartBase = cartsCreated || 1;
  const checkoutBase = abandonedCount || 1;
  const sentBase = whatsapp.sent || 1;
  const deliveredBase = whatsapp.delivered || 1;
  const readBase = whatsapp.read || 1;
  const averageCartValue =
    numberValue(analytics.averageCartValue) ||
    percentage(numberValue(analytics.totalAbandonedRevenue), abandonedCount);
  const timeline = analytics.revenueTimeline || [];
  const statusBreakdown = analytics.statusBreakdown || [];
  const pieData = statusBreakdown.map((item) => ({
    ...item,
    name: item.status,
    value: item.count,
  }));
  const trendData = timeline.map((item) => ({
    name: dateLabel(item.date),
    abandoned: numberValue(item.abandonedCount),
    recovered: numberValue(item.recoveredCount),
  }));
  const funnel = [
    {
      label: "Added to cart",
      value: cartsCreated,
      percent: 100,
      detail: "Shopify cart records created",
    },
    {
      label: "Checkout started",
      value: abandonedCount,
      percent: percentage(abandonedCount, cartBase),
      detail: "Checkout records received",
    },
    {
      label: "WhatsApp sent",
      value: whatsapp.sent,
      percent: percentage(whatsapp.sent, checkoutBase),
      detail: "Recovery message sent or delivered",
    },
    {
      label: "Message delivered",
      value: whatsapp.delivered,
      percent: percentage(whatsapp.delivered, sentBase),
      detail: "Delivered or read",
    },
    {
      label: "Message opened",
      value: whatsapp.read,
      percent: percentage(whatsapp.read, deliveredBase),
      detail: "WhatsApp read status",
    },
    {
      label: "Checkout link clicked",
      value: whatsapp.clicked,
      percent: percentage(whatsapp.clicked, readBase),
      detail: "Tracked recovery click estimate",
    },
    {
      label: "Order completed",
      value: recoveredCount,
      percent: percentage(recoveredCount, checkoutBase),
      detail: "Recovered checkout",
    },
  ];
  const channelRows = [
    {
      label: "Sent",
      value: whatsapp.sent,
      rate: percentage(whatsapp.sent, checkoutBase),
    },
    {
      label: "Delivered",
      value: whatsapp.delivered,
      rate: percentage(whatsapp.delivered, sentBase),
    },
    {
      label: "Read",
      value: whatsapp.read,
      rate: percentage(whatsapp.read, deliveredBase),
    },
    {
      label: "Clicked",
      value: whatsapp.clicked,
      rate: percentage(whatsapp.clicked, readBase),
    },
    {
      label: "Failed",
      value: whatsapp.failed,
      rate: percentage(whatsapp.failed, sentBase),
    },
  ];
  const topLostCarts = analytics.topLostCarts || [];

  return (
    <div className="recovery-analytics">
      <div className="report-metrics-grid abandoned-kpi-grid">
        {[
          {
            label: "Abandoned revenue",
            value: money(analytics.totalAbandonedRevenue),
            detail: "Total checkout value",
          },
          {
            label: "Recovered revenue",
            value: money(analytics.recoveredRevenue),
            detail: `${numberValue(analytics.recoveryRate).toFixed(1)}% recovery rate`,
          },
          {
            label: "Pending value",
            value: money(analytics.pendingRevenue),
            detail: "Value still at risk",
          },
          {
            label: "Carts created",
            value: cartsCreated.toLocaleString("en-IN"),
            detail: "Add-to-cart count",
          },
          {
            label: "Cart → checkout",
            value: `${numberValue(analytics.addCartToCheckoutRate).toFixed(1)}%`,
            detail: `${abandonedCount} checkouts`,
          },
          {
            label: "Cart → order",
            value: `${numberValue(analytics.addCartToOrderRate).toFixed(1)}%`,
            detail: `${recoveredCount} recovered orders`,
          },
          {
            label: "Abandoned carts",
            value: abandonedCount.toLocaleString("en-IN"),
            detail: "Checkout drop-offs",
          },
          {
            label: "Recovered carts",
            value: recoveredCount.toLocaleString("en-IN"),
            detail: "Converted back to order",
          },
          {
            label: "Average cart value",
            value: money(averageCartValue),
            detail: "Average checkout value",
          },
        ].map((metric) => (
          <article className="report-metric-card" key={metric.label}>
            <span className="metric-label">{metric.label}</span>
            <strong>{metric.value}</strong>
            <small>{metric.detail}</small>
          </article>
        ))}
      </div>

      <div className="analytics-grid abandoned-analytics-primary-grid">
        <section className="reports-table-card abandoned-chart-card">
          <div className="reports-table-heading">
            <div>
              <p className="eyebrow">Recovery timeline</p>
              <h3>Abandoned vs recovered value</h3>
            </div>
          </div>
          {timeline.length ? (
            <ResponsiveContainer width="100%" height={260}>
              <AreaChart
                data={timeline}
                margin={{ top: 10, right: 8, left: -18, bottom: 0 }}
              >
                <defs>
                  <linearGradient
                    id="abandonedRevenueFill"
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop offset="5%" stopColor="#ef4444" stopOpacity={0.22} />
                    <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient
                    id="recoveredRevenueFill"
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.22} />
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid vertical={false} stroke="var(--chart-grid)" />
                <XAxis
                  dataKey="date"
                  tickFormatter={dateLabel}
                  stroke="var(--chart-axis)"
                  fontSize={10}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  stroke="var(--chart-axis)"
                  fontSize={10}
                  tickLine={false}
                  axisLine={false}
                />
                <Tooltip />
                <Area
                  type="monotone"
                  dataKey="abandonedAmount"
                  stroke="#ef4444"
                  strokeWidth={2}
                  fill="url(#abandonedRevenueFill)"
                  name="Abandoned"
                />
                <Area
                  type="monotone"
                  dataKey="recoveredAmount"
                  stroke="#10b981"
                  strokeWidth={2}
                  fill="url(#recoveredRevenueFill)"
                  name="Recovered"
                />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <p className="table-state">No timeline data for this period.</p>
          )}
          <div className="orders-table-wrap abandoned-timeline-table">
            <table className="orders-table">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Carts</th>
                  <th>Abandoned</th>
                  <th>Recovered</th>
                </tr>
              </thead>
              <tbody>
                {timeline.length ? (
                  timeline.map((row) => (
                    <tr key={row.date}>
                      <td>{dateLabel(row.date)}</td>
                      <td>{numberValue(row.abandonedCount)}</td>
                      <td className="table-money">
                        {money(row.abandonedAmount)}
                      </td>
                      <td className="table-money">
                        {money(row.recoveredAmount)}
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td className="table-state" colSpan={4}>
                      No timeline data for this period.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </section>

        <section className="reports-table-card abandoned-status-card">
          <div className="reports-table-heading">
            <div>
              <p className="eyebrow">Checkout states</p>
              <h3>Recovery status breakdown</h3>
            </div>
          </div>
          {pieData.length ? (
            <div className="abandoned-pie-layout">
              <ResponsiveContainer width="46%" height={220}>
                <PieChart>
                  <Pie
                    data={pieData}
                    innerRadius={56}
                    outerRadius={82}
                    paddingAngle={3}
                    dataKey="value"
                  >
                    {pieData.map((entry) => (
                      <Cell
                        key={entry.name}
                        fill={statusColors[entry.name] || "#888888"}
                      />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
              <div className="abandoned-status-legend">
                {pieData.map((entry) => (
                  <div className="abandoned-status-legend-row" key={entry.name}>
                    <span
                      className="abandoned-status-dot"
                      style={{
                        background: statusColors[entry.name] || "#888888",
                      }}
                    />
                    <span>
                      <strong>{entry.name}</strong>
                      <small>
                        {entry.count} carts ·{" "}
                        {percentage(entry.count, cartBase).toFixed(0)}% ·{" "}
                        {money(entry.amount)}
                      </small>
                    </span>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <p className="table-state">No status data for this period.</p>
          )}
        </section>
      </div>

      <div className="analytics-grid abandoned-analytics-secondary-grid">
        <section className="reports-table-card abandoned-funnel-card">
          <div className="reports-table-heading">
            <div>
              <p className="eyebrow">Conversion journey</p>
              <h3>Cart &amp; recovery funnel</h3>
            </div>
          </div>
          <div className="abandoned-funnel-list">
            {funnel.map((step) => (
              <div className="abandoned-funnel-step" key={step.label}>
                <div className="abandoned-funnel-label">
                  <span>{step.label}</span>
                  <strong>
                    {step.value.toLocaleString("en-IN")} ·{" "}
                    {Math.min(100, Math.max(0, step.percent)).toFixed(1)}%
                  </strong>
                </div>
                <div className="abandoned-funnel-track">
                  <span
                    style={{
                      width: `${Math.min(100, Math.max(step.percent, step.value > 0 ? 2 : 0))}%`,
                    }}
                  />
                </div>
                <small>{step.detail}</small>
              </div>
            ))}
          </div>
        </section>

        <div className="abandoned-analytics-stack">
          <section className="reports-table-card abandoned-chart-card abandoned-trend-card">
            <div className="reports-table-heading">
              <div>
                <p className="eyebrow">Daily volume</p>
                <h3>Cart abandonment trend</h3>
              </div>
            </div>
            {trendData.length ? (
              <ResponsiveContainer width="100%" height={170}>
                <BarChart
                  data={trendData}
                  margin={{ top: 5, right: 4, left: -18, bottom: 0 }}
                >
                  <CartesianGrid vertical={false} stroke="var(--chart-grid)" />
                  <XAxis
                    dataKey="name"
                    stroke="var(--chart-axis)"
                    fontSize={10}
                    tickLine={false}
                    axisLine={false}
                  />
                  <YAxis
                    allowDecimals={false}
                    stroke="var(--chart-axis)"
                    fontSize={10}
                    tickLine={false}
                    axisLine={false}
                  />
                  <Tooltip />
                  <Bar
                    dataKey="abandoned"
                    fill="var(--status-warning)"
                    radius={[4, 4, 0, 0]}
                    name="Abandoned"
                  />
                  <Bar
                    dataKey="recovered"
                    fill="var(--status-active)"
                    radius={[4, 4, 0, 0]}
                    name="Recovered"
                  />
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <p className="table-state">No daily cart data for this period.</p>
            )}
          </section>

          <section className="reports-table-card abandoned-channel-card">
            <div className="reports-table-heading">
              <div>
                <p className="eyebrow">Automation performance</p>
                <h3>WhatsApp message funnel</h3>
              </div>
            </div>
            <div className="abandoned-channel-summary">
              {channelRows.map((row) => (
                <div key={row.label}>
                  <span>{row.label}</span>
                  <strong>{row.value.toLocaleString("en-IN")}</strong>
                  <small>{row.rate.toFixed(1)}%</small>
                </div>
              ))}
            </div>
          </section>
        </div>
      </div>

      <section className="reports-table-card abandoned-insights-card">
        <div className="reports-table-heading">
          <div>
            <p className="eyebrow">Customer recovery insights</p>
            <h3>Actionable contact and recovery metrics</h3>
          </div>
        </div>
        <div className="abandoned-insights-grid">
          <Insight
            title="Average recovery time"
            value={formatDuration(analytics.averageRecoveryTimeMinutes)}
            detail="Time from cart abandonment to recovered order"
          />
          <Insight
            title="Customer contactability"
            value={`${numberValue(analytics.contactabilityRate).toFixed(1)}%`}
            detail={`${numberValue(analytics.contactableCartCount)} checkouts have a phone number`}
          />
          <Insight
            title="Average attempts to recover"
            value={numberValue(analytics.averageAttemptsToRecovery).toFixed(1)}
            detail="Recovery messages per recovered checkout"
          />
          <Insight
            title="Consent coverage"
            value={`${numberValue(analytics.marketingConsentCount)} marketing · ${numberValue(analytics.smsConsentCount)} SMS`}
            detail="Consent flags captured on checkout records"
          />
        </div>
      </section>

      <section className="reports-table-card abandoned-lost-carts-card">
        <div className="reports-table-heading">
          <div>
            <p className="eyebrow">Largest opportunities</p>
            <h3>Top outstanding lost carts</h3>
            <p>
              Highest-value unrecovered checkouts prioritised for manual
              recovery.
            </p>
          </div>
        </div>
        <div className="orders-table-wrap">
          <table className="orders-table abandoned-lost-carts-table">
            <thead>
              <tr>
                <th>Customer</th>
                <th>Phone</th>
                <th>Cart value</th>
                <th>Abandoned</th>
                <th>Status</th>
                <th>Attempts</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {topLostCarts.length ? (
                topLostCarts.map((cart, index) => {
                  const attempts = cart.recovery_attempts ?? cart.attempts ?? 0;
                  return (
                    <tr key={`${cart.id || cart.customer_name}-${index}`}>
                      <td>{cart.customer_name || "Anonymous customer"}</td>
                      <td>{formatPhone(cart.phone)}</td>
                      <td className="table-money">
                        {money(cart.total_price, cart.currency)}
                      </td>
                      <td>{dateTime(cart.abandoned_at)}</td>
                      <td>
                        <span className="status-pill status-pill-neutral">
                          {cart.recovery_status.replace("_", " ")}
                        </span>
                      </td>
                      <td>{attempts}</td>
                      <td>
                        {cart.id ? (
                          <button
                            className="table-link-button"
                            type="button"
                            onClick={() =>
                              onRecover(cart.id as number, cart.phone)
                            }
                            disabled={sending || !cart.phone}
                          >
                            <Send size={13} aria-hidden="true" /> Recover
                          </button>
                        ) : (
                          "—"
                        )}
                      </td>
                    </tr>
                  );
                })
              ) : (
                <tr>
                  <td className="table-state" colSpan={7}>
                    No outstanding lost carts found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      <p className="communication-footnote">
        <Check size={14} aria-hidden="true" /> Metrics are calculated from cart,
        checkout, recovery, consent, and WhatsApp message records for the
        selected period.
      </p>
    </div>
  );
}

function Insight({
  title,
  value,
  detail,
}: {
  title: string;
  value: string;
  detail: string;
}) {
  return (
    <article className="abandoned-insight">
      <span>{title}</span>
      <strong>{value}</strong>
      <small>{detail}</small>
    </article>
  );
}
