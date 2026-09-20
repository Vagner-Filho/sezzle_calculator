import { useEffect } from 'react';

interface ToastProps {
  message: string;
  onClose: () => void;
}

/**
 * Displays a temporary toast notification for errors.
 */
export function Toast({ message, onClose }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(onClose, 5000);
    return () => clearTimeout(timer);
  }, [onClose]);

  return (
    <div className="toast" role="alert" aria-live="assertive">
      {message}
      <button onClick={onClose} aria-label="Dismiss error" className="toast-close">
        ×
      </button>
    </div>
  );
}
