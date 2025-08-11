import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useRouter } from 'next/navigation';
import { LoginForm } from '../LoginForm';
import { useAuth } from '../../../contexts/AuthContext';

// Mock the dependencies
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

jest.mock('../../../contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

const mockPush = jest.fn();
const mockLogin = jest.fn();

beforeEach(() => {
  (useRouter as jest.Mock).mockReturnValue({
    push: mockPush,
  });
  
  (useAuth as jest.Mock).mockReturnValue({
    login: mockLogin,
  });
  
  jest.clearAllMocks();
});

describe('LoginForm', () => {
  it('renders login form with all required fields', () => {
    render(<LoginForm />);
    
    expect(screen.getByText('Login to Your Account')).toBeInTheDocument();
    expect(screen.getByLabelText('Email Address')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Login' })).toBeInTheDocument();
    expect(screen.getByText('Forgot your password?')).toBeInTheDocument();
    expect(screen.getByText('Sign up')).toBeInTheDocument();
  });

  it('validates required fields', async () => {
    render(<LoginForm />);
    
    const submitButton = screen.getByRole('button', { name: 'Login' });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Email is required')).toBeInTheDocument();
      expect(screen.getByText('Password is required')).toBeInTheDocument();
    });
    
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('validates email format', async () => {
    render(<LoginForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Login' });
    
    fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Email is invalid')).toBeInTheDocument();
    });
    
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('validates password length', async () => {
    render(<LoginForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByRole('button', { name: 'Login' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: '123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Password must be at least 6 characters')).toBeInTheDocument();
    });
    
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('submits form with valid data', async () => {
    mockLogin.mockResolvedValue(undefined);
    
    render(<LoginForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByRole('button', { name: 'Login' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
      });
    });
    
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('handles login error', async () => {
    const errorMessage = 'Invalid credentials';
    mockLogin.mockRejectedValue(new Error(errorMessage));
    
    render(<LoginForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByRole('button', { name: 'Login' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('shows loading state during submission', async () => {
    mockLogin.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));
    
    render(<LoginForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByRole('button', { name: 'Login' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);
    
    expect(screen.getByText('Logging in...')).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
    
    await waitFor(() => {
      expect(screen.getByText('Login')).toBeInTheDocument();
    });
  });

  it('clears errors when user starts typing', async () => {
    render(<LoginForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Login' });
    
    // Trigger validation error
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Email is required')).toBeInTheDocument();
    });
    
    // Start typing to clear error
    fireEvent.change(emailInput, { target: { value: 'test' } });
    
    expect(screen.queryByText('Email is required')).not.toBeInTheDocument();
  });

  it('redirects to custom path when provided', async () => {
    mockLogin.mockResolvedValue(undefined);
    
    render(<LoginForm redirectTo="/custom-path" />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByRole('button', { name: 'Login' });
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/custom-path');
    });
  });
});