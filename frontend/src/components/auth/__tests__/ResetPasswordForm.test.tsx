import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useRouter } from 'next/navigation';
import { ResetPasswordForm } from '../ResetPasswordForm';
import { apiClient } from '../../../lib/api';

// Mock the dependencies
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

jest.mock('../../../lib/api', () => ({
  apiClient: {
    resetPassword: jest.fn(),
  },
}));

const mockPush = jest.fn();

beforeEach(() => {
  (useRouter as jest.Mock).mockReturnValue({
    push: mockPush,
  });
  
  jest.clearAllMocks();
});

describe('ResetPasswordForm', () => {
  const mockToken = 'test-reset-token';

  it('renders reset password form with all required elements', () => {
    render(<ResetPasswordForm token={mockToken} />);
    
    expect(screen.getByText('Set New Password')).toBeInTheDocument();
    expect(screen.getByText('Enter your new password below.')).toBeInTheDocument();
    expect(screen.getByLabelText('New Password')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm New Password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Reset Password' })).toBeInTheDocument();
    expect(screen.getByText('Back to Login')).toBeInTheDocument();
  });

  it('validates required fields', async () => {
    render(<ResetPasswordForm token={mockToken} />);
    
    const submitButton = screen.getByRole('button', { name: 'Reset Password' });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Password is required')).toBeInTheDocument();
      expect(screen.getByText('Please confirm your password')).toBeInTheDocument();
    });
    
    expect(apiClient.resetPassword).not.toHaveBeenCalled();
  });

  it('validates password length', async () => {
    render(<ResetPasswordForm token={mockToken} />);
    
    const newPasswordInput = screen.getByLabelText('New Password');
    const submitButton = screen.getByRole('button', { name: 'Reset Password' });
    
    fireEvent.change(newPasswordInput, { target: { value: '123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Password must be at least 6 characters')).toBeInTheDocument();
    });
    
    expect(apiClient.resetPassword).not.toHaveBeenCalled();
  });

  it('validates password confirmation', async () => {
    render(<ResetPasswordForm token={mockToken} />);
    
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm New Password');
    const submitButton = screen.getByRole('button', { name: 'Reset Password' });
    
    fireEvent.change(newPasswordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'different123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
    });
    
    expect(apiClient.resetPassword).not.toHaveBeenCalled();
  });

  it('submits form with valid data', async () => {
    (apiClient.resetPassword as jest.Mock).mockResolvedValue(undefined);
    
    render(<ResetPasswordForm token={mockToken} />);
    
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm New Password');
    const submitButton = screen.getByRole('button', { name: 'Reset Password' });
    
    fireEvent.change(newPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(apiClient.resetPassword).toHaveBeenCalledWith({
        token: mockToken,
        newPassword: 'newpassword123',
      });
    });
  });

  it('shows success message after successful reset', async () => {
    (apiClient.resetPassword as jest.Mock).mockResolvedValue(undefined);
    
    render(<ResetPasswordForm token={mockToken} />);
    
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm New Password');
    const submitButton = screen.getByRole('button', { name: 'Reset Password' });
    
    fireEvent.change(newPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Password Reset Successful')).toBeInTheDocument();
      expect(screen.getByText('Your password has been successfully reset. You can now login with your new password.')).toBeInTheDocument();
      expect(screen.getByText('Login Now')).toBeInTheDocument();
    });
  });

  it('handles reset password error', async () => {
    const errorMessage = 'Invalid or expired token';
    (apiClient.resetPassword as jest.Mock).mockRejectedValue(new Error(errorMessage));
    
    render(<ResetPasswordForm token={mockToken} />);
    
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm New Password');
    const submitButton = screen.getByRole('button', { name: 'Reset Password' });
    
    fireEvent.change(newPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('shows loading state during submission', async () => {
    (apiClient.resetPassword as jest.Mock).mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 100))
    );
    
    render(<ResetPasswordForm token={mockToken} />);
    
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm New Password');
    const submitButton = screen.getByRole('button', { name: 'Reset Password' });
    
    fireEvent.change(newPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.click(submitButton);
    
    expect(screen.getByText('Resetting Password...')).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
    
    await waitFor(() => {
      expect(screen.getByText('Password Reset Successful')).toBeInTheDocument();
    });
  });

  it('clears errors when user starts typing', async () => {
    render(<ResetPasswordForm token={mockToken} />);
    
    const newPasswordInput = screen.getByLabelText('New Password');
    const submitButton = screen.getByRole('button', { name: 'Reset Password' });
    
    // Trigger validation error
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Password is required')).toBeInTheDocument();
    });
    
    // Start typing to clear error
    fireEvent.change(newPasswordInput, { target: { value: 'test' } });
    
    expect(screen.queryByText('Password is required')).not.toBeInTheDocument();
  });
});