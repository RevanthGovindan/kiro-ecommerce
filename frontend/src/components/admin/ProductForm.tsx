'use client';

import React, { useState } from 'react';
import { apiClient, Product, Category, CreateProductRequest, UpdateProductRequest } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';

interface ProductFormProps {
  product?: Product | null;
  categories: Category[];
  onClose: () => void;
}

export const ProductForm: React.FC<ProductFormProps> = ({
  product,
  categories,
  onClose,
}) => {
  const [formData, setFormData] = useState({
    name: product?.name || '',
    description: product?.description || '',
    price: product?.price?.toString() || '',
    compareAtPrice: product?.compareAtPrice?.toString() || '',
    sku: product?.sku || '',
    inventory: product?.inventory?.toString() || '',
    categoryId: product?.categoryId || '',
    images: product?.images?.join(', ') || '',
    seoTitle: product?.seoTitle || '',
    seoDescription: product?.seoDescription || '',
    isActive: product?.isActive ?? true,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const productData = {
        name: formData.name,
        description: formData.description,
        price: parseFloat(formData.price),
        compareAtPrice: formData.compareAtPrice ? parseFloat(formData.compareAtPrice) : undefined,
        sku: formData.sku,
        inventory: parseInt(formData.inventory),
        categoryId: formData.categoryId,
        images: formData.images.split(',').map(img => img.trim()).filter(img => img),
        seoTitle: formData.seoTitle || undefined,
        seoDescription: formData.seoDescription || undefined,
        isActive: formData.isActive,
      };

      if (product) {
        await apiClient.updateProduct(product.id, productData as UpdateProductRequest);
      } else {
        await apiClient.createProduct(productData as CreateProductRequest);
      }

      onClose();
    } catch (err) {
      setError(product ? 'Failed to update product' : 'Failed to create product');
      console.error('Error saving product:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
    }));
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-medium text-gray-900">
            {product ? 'Edit Product' : 'Add New Product'}
          </h2>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4">
              <p className="text-red-800">{error}</p>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Product Name *"
              name="name"
              value={formData.name}
              onChange={handleChange}
              required
            />

            <Input
              label="SKU *"
              name="sku"
              value={formData.sku}
              onChange={handleChange}
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description *
            </label>
            <textarea
              name="description"
              value={formData.description}
              onChange={handleChange}
              rows={3}
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Input
              label="Price *"
              name="price"
              type="number"
              step="0.01"
              value={formData.price}
              onChange={handleChange}
              required
            />

            <Input
              label="Compare At Price"
              name="compareAtPrice"
              type="number"
              step="0.01"
              value={formData.compareAtPrice}
              onChange={handleChange}
            />

            <Input
              label="Inventory *"
              name="inventory"
              type="number"
              value={formData.inventory}
              onChange={handleChange}
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Category *
            </label>
            <select
              name="categoryId"
              value={formData.categoryId}
              onChange={handleChange}
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">Select a category</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </select>
          </div>

          <Input
            label="Images (comma-separated URLs)"
            name="images"
            value={formData.images}
            onChange={handleChange}
            placeholder="https://example.com/image1.jpg, https://example.com/image2.jpg"
          />

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="SEO Title"
              name="seoTitle"
              value={formData.seoTitle}
              onChange={handleChange}
            />

            <Input
              label="SEO Description"
              name="seoDescription"
              value={formData.seoDescription}
              onChange={handleChange}
            />
          </div>

          <div className="flex items-center">
            <input
              type="checkbox"
              name="isActive"
              checked={formData.isActive}
              onChange={handleChange}
              className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
            />
            <label className="ml-2 block text-sm text-gray-900">
              Product is active
            </label>
          </div>

          <div className="flex justify-end space-x-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Saving...' : (product ? 'Update Product' : 'Create Product')}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};