import React from 'react';
import { render, screen } from '@testing-library/react';
import { HeroSection } from '../HeroSection';

// Mock Next.js Link component
jest.mock('next/link', () => {
  const MockLink = ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  );
  MockLink.displayName = 'MockLink';
  return MockLink;
});

describe('HeroSection', () => {
  it('renders hero content correctly', () => {
    render(<HeroSection />);

    expect(screen.getByText('Welcome to Our Store')).toBeInTheDocument();
    expect(screen.getByText(/Discover amazing products at great prices/)).toBeInTheDocument();
  });

  it('renders call-to-action buttons', () => {
    render(<HeroSection />);

    const shopNowButton = screen.getByRole('link', { name: /shop now/i });
    expect(shopNowButton).toBeInTheDocument();
    expect(shopNowButton).toHaveAttribute('href', '/products');

    const newArrivalsButton = screen.getByRole('link', { name: /view new arrivals/i });
    expect(newArrivalsButton).toBeInTheDocument();
    expect(newArrivalsButton).toHaveAttribute('href', '/products?sort=newest');
  });

  it('has proper styling classes', () => {
    render(<HeroSection />);

    const heroSection = document.querySelector('section');
    expect(heroSection).toHaveClass('bg-gradient-to-r', 'from-blue-600', 'to-blue-800');
  });
});