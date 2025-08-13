'use client';

import React from 'react';
import { ErrorFallbackProps } from './ErrorBoundary';

export const ErrorFallback: React.FC<ErrorFallbackProps> = ({
  error,
  errorInfo,
  resetError,
}) => {
  const isDevelopment = process.env.NODE_ENV === 'development';

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          <div className="text-center">
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
              <svg
                className="h-6 w-6 text-red-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
                />
              </svg>
            </div>
            <h2 className="mt-6 text-2xl font-bold text-gray-900">
              Oops! Something went wrong
            </h2>
            <p className="mt-2 text-sm text-gray-600">
              We're sorry, but something unexpected happened. Our team has been notified.
            </p>
          </div>

          <div className="mt-6">
            <button
              onClick={resetError}
              className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Try Again
            </button>
          </div>

          <div className="mt-4 text-center">
            <button
              onClick={() => window.location.href = '/'}
              className="text-sm text-indigo-600 hover:text-indigo-500"
            >
              Go back to homepage
            </button>
          </div>

          {isDevelopment && error && (
            <div className="mt-6 p-4 bg-red-50 rounded-md">
              <h3 className="text-sm font-medium text-red-800 mb-2">
                Development Error Details:
              </h3>
              <div className="text-xs text-red-700 font-mono bg-red-100 p-2 rounded overflow-auto max-h-40">
                <div className="mb-2">
                  <strong>Error:</strong> {error.message}
                </div>
                {error.stack && (
                  <div className="mb-2">
                    <strong>Stack:</strong>
                    <pre className="whitespace-pre-wrap">{error.stack}</pre>
                  </div>
                )}
                {errorInfo?.componentStack && (
                  <div>
                    <strong>Component Stack:</strong>
                    <pre className="whitespace-pre-wrap">{errorInfo.componentStack}</pre>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};