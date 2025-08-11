import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useRouter } from 'next/navigation';
import { RegisterForm } from '../RegisterForm';
import { useAuth } from '../../../contexts/AuthContext';

// Mock the dependencies
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

jest.mock('../../../contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

const mockPush = jest.fn();
const mockRegister = jest.fn();

beforeEach(() => {
  (useRouter as jest.Mock).mockReturnValue({
    push: mockPush,
  });
  
  (useAuth as jest.Mock).mockReturnValue({
    register: mockRegister,
  });
  
  jest.clearAllMocks();
});

describe('RegisterForm', () => {
  it('renders registration form with all required fields', () => {
    render(<RegisterForm />);
    
    expect(screen.getByText('Create Your Account')).toBeInTheDocument();
    expect(screen.getByLabelText('First Name')).toBeInTheDocument();
    expect(screen.getByLabelText('Last Name')).toBeInTheDocument();
    expect(screen.getByLabelText('Email Address')).toBeInTheDocument();
    expect(screen.getByLabelText('Phone Number (Optional)')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm Password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Create Account' })).toBeInTheDocument();
    expect(screen.getByText('Login')).toBeInTheDocument();
  });

  it('validates required fields', async () => {
    render(<RegisterForm />);
    
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('First name is required')).toBeInTheDocument();
      expect(screen.getByText('Last name is required')).toBeInTheDocument();
      expect(screen.getByText('Email is required')).toBeInTheDocument();
      expect(screen.getByText('Password is required')).toBeInTheDocument();
      expect(screen.getByText('Please confirm your password')).toBeInTheDocument();
    });
    
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it('validates email format', async () => {
    render(<RegisterForm />);
    
    const emailInput = screen.getByLabelText('Email Address');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Email is invalid')).toBeInTheDocument();
    });
    
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it('validates phone number format', async () => {
    render(<RegisterForm />);
    
    const phoneInput = screen.getByLabelText('Phone Number (Optional)');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    fireEvent.change(phoneInput, { target: { value: 'invalid-phone' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Phone number is invalid')).toBeInTheDocument();
    });
    
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it('validates password length', async () => {
    render(<RegisterForm />);
    
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    fireEvent.change(passwordInput, { target: { value: '123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Password must be at least 6 characters')).toBeInTheDocument();
    });
    
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it('validates password confirmation', async () => {
    render(<RegisterForm />);
    
    const passwordInput = screen.getByLabelText('Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'different123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
    });
    
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it('submits form with valid data including phone', async () => {
    mockRegister.mockResolvedValue(undefined);
    
    render(<RegisterForm />);
    
    const firstNameInput = screen.getByLabelText('First Name');
    const lastNameInput = screen.getByLabelText('Last Name');
    const emailInput = screen.getByLabelText('Email Address');
    const phoneInput = screen.getByLabelText('Phone Number (Optional)');
    const passwordInput = screen.getByLabelText('Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    fireEvent.change(firstNameInput, { target: { value: 'John' } });
    fireEvent.change(lastNameInput, { target: { value: 'Doe' } });
    fireEvent.change(emailInput, { target: { value: 'john@example.com' } });
    fireEvent.change(phoneInput, { target: { value: '+1234567890' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        firstName: 'John',
        lastName: 'Doe',
        email: 'john@example.com',
        phone: '+1234567890',
        password: 'password123',
      });
    });
    
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('submits form with valid data without phone', async () => {
    mockRegister.mockResolvedValue(undefined);
    
    render(<RegisterForm />);
    
    const firstNameInput = screen.getByLabelText('First Name');
    const lastNameInput = screen.getByLabelText('Last Name');
    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    fireEvent.change(firstNameInput, { target: { value: 'John' } });
    fireEvent.change(lastNameInput, { target: { value: 'Doe' } });
    fireEvent.change(emailInput, { target: { value: 'john@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        firstName: 'John',
        lastName: 'Doe',
        email: 'john@example.com',
        phone: undefined,
        password: 'password123',
      });
    });
    
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('handles registration error', async () => {
    const errorMessage = 'Email already exists';
    mockRegister.mockRejectedValue(new Error(errorMessage));
    
    render(<RegisterForm />);
    
    const firstNameInput = screen.getByLabelText('First Name');
    const lastNameInput = screen.getByLabelText('Last Name');
    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    fireEvent.change(firstNameInput, { target: { value: 'John' } });
    fireEvent.change(lastNameInput, { target: { value: 'Doe' } });
    fireEvent.change(emailInput, { target: { value: 'john@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('shows loading state during submission', async () => {
    mockRegister.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));
    
    render(<RegisterForm />);
    
    const firstNameInput = screen.getByLabelText('First Name');
    const lastNameInput = screen.getByLabelText('Last Name');
    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    fireEvent.change(firstNameInput, { target: { value: 'John' } });
    fireEvent.change(lastNameInput, { target: { value: 'Doe' } });
    fireEvent.change(emailInput, { target: { value: 'john@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);
    
    expect(screen.getByText('Creating Account...')).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
    
    await waitFor(() => {
      expect(screen.getByText('Create Account')).toBeInTheDocument();
    });
  });

  it('clears errors when user starts typing', async () => {
    render(<RegisterForm />);
    
    const firstNameInput = screen.getByLabelText('First Name');
    const submitButton = screen.getByRole('button', { name: 'Create Account' });
    
    // Trigger validation error
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('First name is required')).toBeInTheDocument();
    });
    
    // Start typing to clear error
    fireEvent.change(firstNameInput, { target: { value: 'John' } });
    
    expect(screen.queryByText('First name is required')).not.toBeInTheDocument();
  });
});