'use client';

import React from 'react';
import { usePathname } from 'next/navigation';
import { Header } from './Header';

export const ConditionalHeader: React.FC = () => {
  const pathname = usePathname();
  
  // Don't show header on admin pages
  if (pathname?.startsWith('/admin')) {
    return null;
  }
  
  return <Header />;
};