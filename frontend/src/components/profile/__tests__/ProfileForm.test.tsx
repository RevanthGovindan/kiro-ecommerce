import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ProfileForm } from '../ProfileForm';
import { useAuth } from '../../../contexts/AuthContext';
import { apiClient } from '../../../lib/api';

// Mock the dependencies
jest.mock('../../../contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

jest.mock('../../../lib/api', () => ({
  apiClient: {
    updateProfile: jest.fn(),
    changePassword: jest.fn(),
  },
}));

const mockUpdateUser = jest.fn();

const mockUser = {
  id: '1',
  email: 'test@example.com',
  firstName: 'John',
  lastName: 'Doe',
  phone: '+1234567890',
  role: 'customer',
  isActive: true,
  createdAt: '2023-01-01T00:00:00Z',
  updatedAt: '2023-01-01T00:00:00Z',
};

beforeEach(() => {
  (useAuth as jest.Mock).mockReturnValue({
    user: mockUser,
    updateUser: mockUpdateUser,
  });
  
  jest.clearAllMocks();
});

describe('ProfileForm', () => {
  it('renders profile form with user data', () => {
    render(<ProfileForm />);
    
    expect(screen.getByText('Profile Information')).toBeInTheDocument();
    expect(screen.getByText('Change Password')).toBeInTheDocument();
    expect(screen.getByDisplayValue('John')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Doe')).toBeInTheDocument();
    expect(screen.getByDisplayValue('+1234567890')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test@example.com')).toBeInTheDocument();
  });

  it('shows message when user is not logged in', () => {
    (useAuth as jest.Mock).mockReturnValue({
      user: null,
      updateUser: mockUpdateUser,
    });
    
    render(<ProfileForm />);
    
    expect(screen.getByText('Please log in to view your profile.')).toBeInTheDocument();
  });

  it('validates required fields in profile form', async () => {
    render(<ProfileForm />);
    
    const firstNameInput = screen.getByLabelText('First Name');
    const lastNameInput = screen.getByLabelText('Last Name');
    const submitButton = screen.getByRole('button', { name: 'Update Profile' });
    
    // Clear required fields
    fireEvent.change(firstNameInput, { target: { value: '' } });
    fireEvent.change(lastNameInput, { target: { value: '' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('First name is required')).toBeInTheDocument();
      expect(screen.getByText('Last name is required')).toBeInTheDocument();
    });
    
    expect(apiClient.updateProfile).not.toHaveBeenCalled();
  });

  it('validates phone number format', async () => {
    render(<ProfileForm />);
    
    const phoneInput = screen.getByLabelText('Phone Number');
    const submitButton = screen.getByRole('button', { name: 'Update Profile' });
    
    fireEvent.change(phoneInput, { target: { value: 'invalid-phone' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Phone number is invalid')).toBeInTheDocument();
    });
    
    expect(apiClient.updateProfile).not.toHaveBeenCalled();
  });

  it('submits profile form with valid data', async () => {
    const updatedUser = { ...mockUser, firstName: 'Jane' };
    (apiClient.updateProfile as jest.Mock).mockResolvedValue(updatedUser);
    
    render(<ProfileForm />);
    
    const firstNameInput = screen.getByLabelText('First Name');
    const submitButton = screen.getByRole('button', { name: 'Update Profile' });
    
    fireEvent.change(firstNameInput, { target: { value: 'Jane' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(apiClient.updateProfile).toHaveBeenCalledWith({
        firstName: 'Jane',
        lastName: 'Doe',
        phone: '+1234567890',
      });
    });
    
    expect(mockUpdateUser).toHaveBeenCalledWith(updatedUser);
    expect(screen.getByText('Profile updated successfully!')).toBeInTheDocument();
  });

  it('handles profile update error', async () => {
    const errorMessage = 'Update failed';
    (apiClient.updateProfile as jest.Mock).mockRejectedValue(new Error(errorMessage));
    
    render(<ProfileForm />);
    
    const submitButton = screen.getByRole('button', { name: 'Update Profile' });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('switches to password tab and validates password fields', async () => {
    render(<ProfileForm />);
    
    // Switch to password tab
    const passwordTab = screen.getByText('Change Password');
    fireEvent.click(passwordTab);
    
    const submitButton = screen.getByRole('button', { name: 'Change Password' });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Current password is required')).toBeInTheDocument();
      expect(screen.getByText('New password is required')).toBeInTheDocument();
      expect(screen.getByText('Please confirm your new password')).toBeInTheDocument();
    });
    
    expect(apiClient.changePassword).not.toHaveBeenCalled();
  });

  it('validates password length and confirmation', async () => {
    render(<ProfileForm />);
    
    // Switch to password tab
    const passwordTab = screen.getByText('Change Password');
    fireEvent.click(passwordTab);
    
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm New Password');
    const submitButton = screen.getByRole('button', { name: 'Change Password' });
    
    fireEvent.change(newPasswordInput, { target: { value: '123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'different' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Password must be at least 6 characters')).toBeInTheDocument();
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
    });
    
    expect(apiClient.changePassword).not.toHaveBeenCalled();
  });

  it('submits password change with valid data', async () => {
    (apiClient.changePassword as jest.Mock).mockResolvedValue(undefined);
    
    render(<ProfileForm />);
    
    // Switch to password tab
    const passwordTab = screen.getByText('Change Password');
    fireEvent.click(passwordTab);
    
    const currentPasswordInput = screen.getByLabelText('Current Password');
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm New Password');
    const submitButton = screen.getByRole('button', { name: 'Change Password' });
    
    fireEvent.change(currentPasswordInput, { target: { value: 'currentpassword' } });
    fireEvent.change(newPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(apiClient.changePassword).toHaveBeenCalledWith({
        currentPassword: 'currentpassword',
        newPassword: 'newpassword123',
      });
    });
    
    expect(screen.getByText('Password changed successfully!')).toBeInTheDocument();
    
    // Form should be cleared
    expect(currentPasswordInput).toHaveValue('');
    expect(newPasswordInput).toHaveValue('');
    expect(confirmPasswordInput).toHaveValue('');
  });

  it('handles password change error', async () => {
    const errorMessage = 'Current password is incorrect';
    (apiClient.changePassword as jest.Mock).mockRejectedValue(new Error(errorMessage));
    
    render(<ProfileForm />);
    
    // Switch to password tab
    const passwordTab = screen.getByText('Change Password');
    fireEvent.click(passwordTab);
    
    const currentPasswordInput = screen.getByLabelText('Current Password');
    const newPasswordInput = screen.getByLabelText('New Password');
    const confirmPasswordInput = screen.getByLabelText('Confirm New Password');
    const submitButton = screen.getByRole('button', { name: 'Change Password' });
    
    fireEvent.change(currentPasswordInput, { target: { value: 'wrongpassword' } });
    fireEvent.change(newPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'newpassword123' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('shows loading states during form submission', async () => {
    (apiClient.updateProfile as jest.Mock).mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 100))
    );
    
    render(<ProfileForm />);
    
    const submitButton = screen.getByRole('button', { name: 'Update Profile' });
    fireEvent.click(submitButton);
    
    expect(screen.getByText('Updating...')).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
    
    await waitFor(() => {
      expect(screen.getByText('Update Profile')).toBeInTheDocument();
    });
  });

  it('clears errors when user starts typing', async () => {
    render(<ProfileForm />);
    
    const firstNameInput = screen.getByLabelText('First Name');
    const submitButton = screen.getByRole('button', { name: 'Update Profile' });
    
    // Clear field to trigger error
    fireEvent.change(firstNameInput, { target: { value: '' } });
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('First name is required')).toBeInTheDocument();
    });
    
    // Start typing to clear error
    fireEvent.change(firstNameInput, { target: { value: 'Jane' } });
    
    expect(screen.queryByText('First name is required')).not.toBeInTheDocument();
  });
});