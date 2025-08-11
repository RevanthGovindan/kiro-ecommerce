'use client';

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { Cart, apiClient } from '@/lib/api';

interface CartContextType {
  cart: Cart | null;
  loading: boolean;
  error: string | null;
  addToCart: (productId: string, quantity?: number) => Promise<void>;
  updateCartItem: (productId: string, quantity: number) => Promise<void>;
  removeFromCart: (productId: string) => Promise<void>;
  clearCart: () => Promise<void>;
  refreshCart: () => Promise<void>;
  getItemCount: () => number;
}

const CartContext = createContext<CartContextType | undefined>(undefined);

export const useCart = () => {
  const context = useContext(CartContext);
  if (context === undefined) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
};

interface CartProviderProps {
  children: ReactNode;
}

export const CartProvider: React.FC<CartProviderProps> = ({ children }) => {
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refreshCart = async () => {
    try {
      setLoading(true);
      setError(null);
      const cartData = await apiClient.getCart();
      setCart(cartData);
    } catch (err) {
      console.error('Error fetching cart:', err);
      setError('Failed to load cart');
      // Initialize empty cart on error
      setCart({
        sessionId: '',
        items: [],
        subtotal: 0,
        tax: 0,
        total: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      });
    } finally {
      setLoading(false);
    }
  };

  const addToCart = async (productId: string, quantity: number = 1) => {
    try {
      setError(null);
      const updatedCart = await apiClient.addToCart(productId, quantity);
      setCart(updatedCart);
    } catch (err) {
      console.error('Error adding to cart:', err);
      setError('Failed to add item to cart');
      throw err;
    }
  };

  const updateCartItem = async (productId: string, quantity: number) => {
    try {
      setError(null);
      const updatedCart = await apiClient.updateCartItem(productId, quantity);
      setCart(updatedCart);
    } catch (err) {
      console.error('Error updating cart item:', err);
      setError('Failed to update cart item');
      throw err;
    }
  };

  const removeFromCart = async (productId: string) => {
    try {
      setError(null);
      const updatedCart = await apiClient.removeFromCart(productId);
      setCart(updatedCart);
    } catch (err) {
      console.error('Error removing from cart:', err);
      setError('Failed to remove item from cart');
      throw err;
    }
  };

  const clearCart = async () => {
    try {
      setError(null);
      await apiClient.clearCart();
      setCart({
        sessionId: cart?.sessionId || '',
        items: [],
        subtotal: 0,
        tax: 0,
        total: 0,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      });
    } catch (err) {
      console.error('Error clearing cart:', err);
      setError('Failed to clear cart');
      throw err;
    }
  };

  const getItemCount = () => {
    if (!cart) return 0;
    return cart.items.reduce((total, item) => total + item.quantity, 0);
  };

  useEffect(() => {
    refreshCart();
  }, []);

  const value: CartContextType = {
    cart,
    loading,
    error,
    addToCart,
    updateCartItem,
    removeFromCart,
    clearCart,
    refreshCart,
    getItemCount,
  };

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
};