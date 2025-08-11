const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  compareAtPrice?: number;
  sku: string;
  inventory: number;
  isActive: boolean;
  categoryId: string;
  images: string[];
  specifications?: Record<string, unknown>;
  seoTitle?: string;
  seoDescription?: string;
  createdAt: string;
  updatedAt: string;
  category?: Category;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  parentId?: string;
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
  children?: Category[];
  parent?: Category;
}

export interface ProductListResponse {
  products: Product[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data: T;
  timestamp: string;
}

export interface CartItem {
  productId: string;
  quantity: number;
  price: number;
  total: number;
  product?: Product;
}

export interface Cart {
  sessionId: string;
  userId?: string;
  items: CartItem[];
  subtotal: number;
  tax: number;
  total: number;
  createdAt: string;
  updatedAt: string;
}

export interface OrderAddress {
  firstName: string;
  lastName: string;
  company?: string;
  address1: string;
  address2?: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
  phone?: string;
}

export interface OrderItem {
  id: string;
  orderId: string;
  productId: string;
  quantity: number;
  price: number;
  total: number;
  product?: Product;
}

export interface Order {
  id: string;
  userId: string;
  status: string;
  subtotal: number;
  tax: number;
  shipping: number;
  total: number;
  shippingAddress: OrderAddress;
  billingAddress: OrderAddress;
  paymentIntentId: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
  items?: OrderItem[];
}

export interface CreateOrderRequest {
  shippingAddress: OrderAddress;
  billingAddress: OrderAddress;
  notes?: string;
}

export interface Payment {
  id: string;
  orderId: string;
  razorpayOrderId: string;
  razorpayPaymentId?: string;
  razorpaySignature?: string;
  amount: number;
  currency: string;
  status: string;
  method?: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  phone?: string;
  role: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  addresses?: Address[];
}

export interface Address {
  id?: string;
  userId?: string;
  type: 'shipping' | 'billing';
  firstName: string;
  lastName: string;
  company?: string;
  address1: string;
  address2?: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
  phone?: string;
  isDefault?: boolean;
}

export interface AuthResponse {
  user: User;
  token: string;
  refreshToken: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  phone?: string;
}

export interface UpdateProfileRequest {
  firstName?: string;
  lastName?: string;
  phone?: string;
}

export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  newPassword: string;
}

// Admin interfaces
export interface AdminProductFilters {
  categoryId?: string;
  isActive?: boolean;
  minPrice?: number;
  maxPrice?: number;
  search?: string;
}

export interface AdminProductListResponse {
  products: Product[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface CreateProductRequest {
  name: string;
  description: string;
  price: number;
  compareAtPrice?: number;
  sku: string;
  inventory: number;
  categoryId: string;
  images: string[];
  specifications?: Record<string, unknown>;
  seoTitle?: string;
  seoDescription?: string;
}

export interface UpdateProductRequest {
  name?: string;
  description?: string;
  price?: number;
  compareAtPrice?: number;
  sku?: string;
  inventory?: number;
  categoryId?: string;
  images?: string[];
  specifications?: Record<string, unknown>;
  seoTitle?: string;
  seoDescription?: string;
  isActive?: boolean;
}

export interface UpdateInventoryRequest {
  inventory: number;
}

export interface AdminOrderFilters {
  status?: string;
  userId?: string;
  startDate?: string;
  endDate?: string;
  search?: string;
}

export interface AdminOrderListResponse {
  orders: Order[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface UpdateOrderStatusRequest {
  status: string;
  notes?: string;
}

export interface AdminCustomerFilters {
  search?: string;
  isActive?: boolean;
}

export interface AdminCustomerListResponse {
  customers: User[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface SalesMetrics {
  totalRevenue: number;
  totalOrders: number;
  averageOrderValue: number;
  topProducts: Array<{
    product: Product;
    totalSold: number;
    revenue: number;
  }>;
  revenueByMonth: Array<{
    month: string;
    revenue: number;
    orders: number;
  }>;
}

// Advanced Search Types
export interface AdvancedSearchResponse {
  products: Product[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  suggestions?: string[];
  facets?: SearchFacets;
}

export interface SearchFacets {
  categories: CategoryFacet[];
  priceRanges: PriceRangeFacet[];
}

export interface CategoryFacet {
  id: string;
  name: string;
  count: number;
}

export interface PriceRangeFacet {
  range: string;
  min: number;
  max: number;
  count: number;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private getAuthHeaders(): Record<string, string> {
    const token = typeof window !== 'undefined' ? localStorage.getItem('authToken') : null;
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...this.getAuthHeaders(),
        ...options.headers,
      },
      ...options,
    });

    if (!response.ok) {
      if (response.status === 401) {
        // Token expired or invalid, clear local storage
        if (typeof window !== 'undefined') {
          localStorage.removeItem('authToken');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('user');
        }
      }
      throw new Error(`API request failed: ${response.status} ${response.statusText}`);
    }

    const result: ApiResponse<T> = await response.json();
    return result.data;
  }

  // Product methods
  async getProducts(params: {
    page?: number;
    pageSize?: number;
    categoryId?: string;
    minPrice?: number;
    maxPrice?: number;
    inStock?: boolean;
    search?: string;
    sortBy?: 'name' | 'price' | 'created_at' | 'updated_at';
    sortOrder?: 'asc' | 'desc';
  } = {}): Promise<ProductListResponse> {
    const searchParams = new URLSearchParams();
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, value.toString());
      }
    });

    return this.request<ProductListResponse>(`/api/products?${searchParams.toString()}`);
  }

  async getProduct(id: string): Promise<Product> {
    return this.request<Product>(`/api/products/${id}`);
  }

  async searchProducts(query: string, page: number = 1, pageSize: number = 20): Promise<ProductListResponse> {
    const searchParams = new URLSearchParams({
      q: query,
      page: page.toString(),
      page_size: pageSize.toString(),
    });

    return this.request<ProductListResponse>(`/api/products/search?${searchParams.toString()}`);
  }

  async advancedSearchProducts(params: {
    query?: string;
    categoryId?: string;
    minPrice?: number;
    maxPrice?: number;
    inStock?: boolean;
    sortBy?: 'name' | 'price' | 'created_at' | 'popularity';
    sortOrder?: 'asc' | 'desc';
    page?: number;
    pageSize?: number;
    includeFacets?: boolean;
  } = {}): Promise<AdvancedSearchResponse> {
    const searchParams = new URLSearchParams();
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        if (key === 'query') {
          searchParams.append('q', value.toString());
        } else if (key === 'pageSize') {
          searchParams.append('page_size', value.toString());
        } else if (key === 'sortBy') {
          searchParams.append('sort_by', value.toString());
        } else if (key === 'sortOrder') {
          searchParams.append('sort_order', value.toString());
        } else if (key === 'includeFacets') {
          searchParams.append('include_facets', value.toString());
        } else {
          searchParams.append(key, value.toString());
        }
      }
    });

    return this.request<AdvancedSearchResponse>(`/api/products/advanced-search?${searchParams.toString()}`);
  }

  async getSearchSuggestions(query: string, size: number = 5): Promise<{ suggestions: string[] }> {
    const searchParams = new URLSearchParams({
      q: query,
      size: size.toString(),
    });

    return this.request<{ suggestions: string[] }>(`/api/products/suggestions?${searchParams.toString()}`);
  }

  // Category methods
  async getCategories(): Promise<{ categories: Category[]; total: number }> {
    return this.request<{ categories: Category[]; total: number }>('/api/categories');
  }

  async getCategory(id: string): Promise<Category> {
    return this.request<Category>(`/api/categories/${id}`);
  }

  // Cart methods
  async getCart(): Promise<Cart> {
    return this.request<Cart>('/api/cart');
  }

  async addToCart(productId: string, quantity: number = 1): Promise<Cart> {
    return this.request<Cart>('/api/cart/add', {
      method: 'POST',
      body: JSON.stringify({ productId, quantity }),
    });
  }

  async updateCartItem(productId: string, quantity: number): Promise<Cart> {
    return this.request<Cart>('/api/cart/update', {
      method: 'PUT',
      body: JSON.stringify({ productId, quantity }),
    });
  }

  async removeFromCart(productId: string): Promise<Cart> {
    return this.request<Cart>('/api/cart/remove', {
      method: 'DELETE',
      body: JSON.stringify({ productId }),
    });
  }

  async clearCart(): Promise<void> {
    return this.request<void>('/api/cart/clear', {
      method: 'DELETE',
    });
  }

  // Order methods
  async createOrder(orderData: CreateOrderRequest): Promise<Order> {
    return this.request<Order>('/api/orders/create', {
      method: 'POST',
      body: JSON.stringify(orderData),
    });
  }

  async getOrder(orderId: string): Promise<Order> {
    return this.request<Order>(`/api/orders/${orderId}`);
  }

  async getUserOrders(page: number = 1, limit: number = 10): Promise<{
    orders: Order[];
    pagination: {
      page: number;
      limit: number;
      total: number;
      totalPages: number;
    };
  }> {
    const searchParams = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });
    return this.request<{
      orders: Order[];
      pagination: {
        page: number;
        limit: number;
        total: number;
        totalPages: number;
      };
    }>(`/api/orders?${searchParams.toString()}`);
  }

  // Payment methods
  async createPaymentOrder(orderId: string, amount: number, currency: string = 'INR'): Promise<Payment> {
    return this.request<Payment>('/api/payments/create-order', {
      method: 'POST',
      body: JSON.stringify({ orderId, amount, currency }),
    });
  }

  async verifyPayment(
    razorpayOrderId: string,
    razorpayPaymentId: string,
    razorpaySignature: string
  ): Promise<void> {
    return this.request<void>('/api/payments/verify', {
      method: 'POST',
      body: JSON.stringify({
        razorpayOrderId,
        razorpayPaymentId,
        razorpaySignature,
      }),
    });
  }

  async getPaymentStatus(orderId: string): Promise<Payment> {
    return this.request<Payment>(`/api/payments/status/${orderId}`);
  }

  // Authentication methods
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  }

  async register(userData: RegisterRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  async logout(): Promise<void> {
    return this.request<void>('/api/auth/logout', {
      method: 'POST',
    });
  }

  async refreshToken(): Promise<AuthResponse> {
    const refreshToken = typeof window !== 'undefined' ? localStorage.getItem('refreshToken') : null;
    return this.request<AuthResponse>('/api/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    });
  }

  async forgotPassword(email: ForgotPasswordRequest): Promise<void> {
    return this.request<void>('/api/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify(email),
    });
  }

  async resetPassword(resetData: ResetPasswordRequest): Promise<void> {
    return this.request<void>('/api/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify(resetData),
    });
  }

  // User profile methods
  async getProfile(): Promise<User> {
    return this.request<User>('/api/users/profile');
  }

  async updateProfile(profileData: UpdateProfileRequest): Promise<User> {
    return this.request<User>('/api/users/profile', {
      method: 'PUT',
      body: JSON.stringify(profileData),
    });
  }

  async changePassword(passwordData: ChangePasswordRequest): Promise<void> {
    return this.request<void>('/api/users/change-password', {
      method: 'PUT',
      body: JSON.stringify(passwordData),
    });
  }

  async getUserAddresses(): Promise<Address[]> {
    return this.request<Address[]>('/api/users/addresses');
  }

  async addAddress(address: Omit<Address, 'id' | 'userId'>): Promise<Address> {
    return this.request<Address>('/api/users/addresses', {
      method: 'POST',
      body: JSON.stringify(address),
    });
  }

  async updateAddress(addressId: string, address: Partial<Address>): Promise<Address> {
    return this.request<Address>(`/api/users/addresses/${addressId}`, {
      method: 'PUT',
      body: JSON.stringify(address),
    });
  }

  async deleteAddress(addressId: string): Promise<void> {
    return this.request<void>(`/api/users/addresses/${addressId}`, {
      method: 'DELETE',
    });
  }

  // Admin methods
  async getAdminProducts(params: {
    page?: number;
    pageSize?: number;
    categoryId?: string;
    isActive?: boolean;
    minPrice?: number;
    maxPrice?: number;
    search?: string;
    sortBy?: 'name' | 'price' | 'created_at' | 'updated_at';
    sortOrder?: 'asc' | 'desc';
  } = {}): Promise<AdminProductListResponse> {
    const searchParams = new URLSearchParams();
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, value.toString());
      }
    });

    return this.request<AdminProductListResponse>(`/api/admin/products?${searchParams.toString()}`);
  }

  async createProduct(productData: CreateProductRequest): Promise<Product> {
    return this.request<Product>('/api/admin/products', {
      method: 'POST',
      body: JSON.stringify(productData),
    });
  }

  async updateProduct(productId: string, productData: UpdateProductRequest): Promise<Product> {
    return this.request<Product>(`/api/admin/products/${productId}`, {
      method: 'PUT',
      body: JSON.stringify(productData),
    });
  }

  async deleteProduct(productId: string): Promise<void> {
    return this.request<void>(`/api/admin/products/${productId}`, {
      method: 'DELETE',
    });
  }

  async updateProductInventory(productId: string, inventory: number): Promise<Product> {
    return this.request<Product>(`/api/admin/products/${productId}/inventory`, {
      method: 'PUT',
      body: JSON.stringify({ inventory }),
    });
  }

  async getAdminOrders(params: {
    page?: number;
    pageSize?: number;
    status?: string;
    userId?: string;
    startDate?: string;
    endDate?: string;
    search?: string;
    sortBy?: 'created_at' | 'updated_at' | 'total';
    sortOrder?: 'asc' | 'desc';
  } = {}): Promise<AdminOrderListResponse> {
    const searchParams = new URLSearchParams();
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, value.toString());
      }
    });

    return this.request<AdminOrderListResponse>(`/api/admin/orders?${searchParams.toString()}`);
  }

  async updateOrderStatus(orderId: string, status: string, notes?: string): Promise<Order> {
    return this.request<Order>(`/api/admin/orders/${orderId}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status, notes }),
    });
  }

  async getAdminCustomers(params: {
    page?: number;
    pageSize?: number;
    search?: string;
    isActive?: boolean;
    sortBy?: 'created_at' | 'updated_at' | 'firstName' | 'lastName';
    sortOrder?: 'asc' | 'desc';
  } = {}): Promise<AdminCustomerListResponse> {
    const searchParams = new URLSearchParams();
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, value.toString());
      }
    });

    return this.request<AdminCustomerListResponse>(`/api/admin/customers?${searchParams.toString()}`);
  }

  async getSalesMetrics(params: {
    startDate?: string;
    endDate?: string;
  } = {}): Promise<SalesMetrics> {
    const searchParams = new URLSearchParams();
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, value.toString());
      }
    });

    return this.request<SalesMetrics>(`/api/admin/analytics/sales?${searchParams.toString()}`);
  }
}

export const apiClient = new ApiClient();