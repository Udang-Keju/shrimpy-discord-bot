"use client";

import { createContext, useCallback, useRef, useState } from "react";
import { CheckCircle2, AlertTriangle, XCircle, Info, Loader2, X } from "lucide-react";
import styles from "./Toast.module.css";

export type ToastVariant = "success" | "warning" | "error" | "info" | "loading";

interface ToastItem {
  id: number;
  message: string;
  variant: ToastVariant;
  exiting: boolean;
}

interface ToastContextValue {
  showToast: (message: string, variant?: ToastVariant) => number;
  updateToast: (id: number, message: string, variant?: ToastVariant) => void;
  dismissToast: (id: number) => void;
}

export const ToastContext = createContext<ToastContextValue | null>(null);

const AUTO_DISMISS_MS = 3500;
const EXIT_ANIMATION_MS = 200;

const ICONS: Record<ToastVariant, typeof CheckCircle2> = {
  success: CheckCircle2,
  warning: AlertTriangle,
  error: XCircle,
  info: Info,
  loading: Loader2,
};

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const nextId = useRef(0);

  const removeToast = useCallback((id: number) => {
    setToasts((prev) =>
      prev.map((t) => (t.id === id ? { ...t, exiting: true } : t))
    );
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, EXIT_ANIMATION_MS);
  }, []);

  const showToast = useCallback(
    (message: string, variant: ToastVariant = "info") => {
      const id = nextId.current++;
      setToasts((prev) => [...prev, { id, message, variant, exiting: false }]);
      if (variant !== "loading") {
        setTimeout(() => removeToast(id), AUTO_DISMISS_MS);
      }
      return id;
    },
    [removeToast]
  );

  const updateToast = useCallback(
    (id: number, message: string, variant: ToastVariant = "info") => {
      setToasts((prev) =>
        prev.map((t) => (t.id === id ? { ...t, message, variant } : t))
      );
      if (variant !== "loading") {
        setTimeout(() => removeToast(id), AUTO_DISMISS_MS);
      }
    },
    [removeToast]
  );

  const dismissToast = useCallback((id: number) => removeToast(id), [removeToast]);

  return (
    <ToastContext.Provider value={{ showToast, updateToast, dismissToast }}>
      {children}
      <div className={styles.container}>
        {toasts.map((toast) => {
          const Icon = ICONS[toast.variant];
          return (
            <div
              key={toast.id}
              className={`${styles.toast} ${styles[toast.variant]} ${
                toast.exiting ? styles.exiting : ""
              }`}
              role="status"
            >
              <Icon size={18} className={styles.icon} />
              <span className={styles.message}>{toast.message}</span>
              <button
                type="button"
                className={styles.closeButton}
                aria-label="Dismiss notification"
                onClick={() => removeToast(toast.id)}
              >
                <X size={16} />
              </button>
            </div>
          );
        })}
      </div>
    </ToastContext.Provider>
  );
}
