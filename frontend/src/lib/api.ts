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

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    });

    if (!response.ok) {
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

  // Category methods
  async getCategories(): Promise<{ categories: Category[]; total: number }> {
    return this.request<{ categories: Category[]; total: number }>('/api/categories');
  }

  async getCategory(id: string): Promise<Category> {
    return this.request<Category>(`/api/categories/${id}`);
  }
}

export const apiClient = new ApiClient();