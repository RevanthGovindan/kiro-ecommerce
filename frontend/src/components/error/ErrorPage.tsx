'use client';

import React from 'react';
import Link from 'next/link';

interface ErrorPageProps {
  statusCode?: number;
  title?: string;
  message?: string;
  showRetry?: boolean;
  onRetry?: () => void;
}

export const ErrorPage: React.FC<ErrorPageProps> = ({
  statusCode = 500,
  title,
  message,
  showRetry = false,
  onRetry,
}) => {
  const getDefaultTitle = (code: number) => {
    switch (code) {
      case 400:
        return 'Bad Request';
      case 401:
        return 'Unauthorized';
      case 403:
        return 'Forbidden';
      case 404:
        return 'Page Not Found';
      case 429:
        return 'Too Many Requests';
      case 500:
        return 'Internal Server Error';
      case 502:
        return 'Bad Gateway';
      case 503:
        return 'Service Unavailable';
      default:
        return 'Something went wrong';
    }
  };

  const getDefaultMessage = (code: number) => {
    switch (code) {
      case 400:
        return 'The request was invalid. Please check your input and try again.';
      case 401:
        return 'You need to sign in to access this page.';
      case 403:
        return 'You don\'t have permission to access this resource.';
      case 404:
        return 'The page you\'re looking for doesn\'t exist.';
      case 429:
        return 'You\'ve made too many requests. Please wait a moment and try again.';
      case 500:
        return 'We\'re experiencing technical difficulties. Please try again later.';
      case 502:
        return 'We\'re having trouble connecting to our servers. Please try again later.';
      case 503:
        return 'Our service is temporarily unavailable. Please try again later.';
      default:
        return 'An unexpected error occurred. Please try again later.';
    }
  };

  const getIcon = (code: number) => {
    if (code === 404) {
      return (
        <svg
          className="h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1}
            d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6-4h6m2 5.291A7.962 7.962 0 0112 15c-2.34 0-4.47-.881-6.08-2.33"
          />
        </svg>
      );
    }

    return (
      <svg
        className="h-12 w-12 text-red-400"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={1}
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
        />
      </svg>
    );
  };

  const displayTitle = title || getDefaultTitle(statusCode);
  const displayMessage = message || getDefaultMessage(statusCode);

  return (
    <div className="min-h-screen bg-white px-4 py-16 sm:px-6 sm:py-24 md:grid md:place-items-center lg:px-8">
      <div className="max-w-max mx-auto">
        <main className="sm:flex">
          <div className="flex-shrink-0 flex justify-center">
            {getIcon(statusCode)}
          </div>
          <div className="sm:ml-6">
            <div className="sm:border-l sm:border-gray-200 sm:pl-6">
              <h1 className="text-4xl font-extrabold text-gray-900 tracking-tight sm:text-5xl">
                {statusCode}
              </h1>
              <p className="mt-1 text-base text-gray-500">{displayTitle}</p>
            </div>
            <div className="mt-8 flex space-x-3 sm:border-l sm:border-transparent sm:pl-6">
              <Link
                href="/"
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Go back home
              </Link>
              {showRetry && onRetry && (
                <button
                  onClick={onRetry}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Try again
                </button>
              )}
              <Link
                href="/contact"
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Contact support
              </Link>
            </div>
            <div className="mt-6">
              <p className="text-base text-gray-500">{displayMessage}</p>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
};