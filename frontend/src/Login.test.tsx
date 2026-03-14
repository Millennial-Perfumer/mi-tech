import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { Login } from './Login';

describe('Login Component', () => {
  it('renders login form correctly', () => {
    render(<Login onLogin={() => {}} />);

    expect(screen.getByText('Welcome Back')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('admin')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('••••••••')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('handles invalid credentials', async () => {
    // Mock fetch
    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: false,
      text: () => Promise.resolve('Invalid credentials')
    });

    render(<Login onLogin={() => {}} />);

    fireEvent.change(screen.getByPlaceholderText('admin'), { target: { value: 'wronguser' } });
    fireEvent.change(screen.getByPlaceholderText('••••••••'), { target: { value: 'wrongpass' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
    });
  });

  it('calls onLogin upon successful login', async () => {
    const mockOnLogin = vi.fn();
    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ token: 'mock_token' })
    });

    render(<Login onLogin={mockOnLogin} />);

    fireEvent.change(screen.getByPlaceholderText('admin'), { target: { value: 'testuser' } });
    fireEvent.change(screen.getByPlaceholderText('••••••••'), { target: { value: 'testpass' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(mockOnLogin).toHaveBeenCalledWith('mock_token');
    });
  });
});
