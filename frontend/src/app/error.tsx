'use client';

import { useEffect } from 'react';
import { ErrorPage } from '@/components/error/ErrorPage';

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.error('Application error:', error);
    
    // Send error to monitoring endpoint
    fetch('/api/errors/client', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        message: error.message,
        stack: error.stack,
        digest: error.digest,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
        url: window.location.href,
      }),
    }).catch((err) => {
      console.error('Failed to log error:', err);
    });
  }, [error]);

  return (
    <ErrorPage
      statusCode={500}
      title="Application Error"
      message="Something went wrong in our application. Our team has been notified and is working on a fix."
      showRetry={true}
      onRetry={reset}
    />
  );
}