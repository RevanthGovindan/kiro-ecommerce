'use client';

import React, { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Image from 'next/image';
import Link from 'next/link';
import { Product, apiClient } from '@/lib/api';

import { Button } from '@/components/ui/Button';

export default function ProductDetailPage() {
  const params = useParams();
  const productId = params?.id as string;
  
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [quantity, setQuantity] = useState(1);

  useEffect(() => {
    const fetchProduct = async () => {
      if (!productId) return;
      
      try {
        setLoading(true);
        setError(null);
        
        const productData = await apiClient.getProduct(productId);
        setProduct(productData);
      } catch (err) {
        console.error('Error fetching product:', err);
        setError('Product not found');
      } finally {
        setLoading(false);
      }
    };

    fetchProduct();
  }, [productId]);



  const handleAddToCart = () => {
    if (!product) return;
    
    console.log('Add to cart:', { productId: product.id, quantity });
    alert(`Added ${quantity} ${product.name}(s) to cart`);
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
              <div className="h-96 bg-gray-200 rounded-lg"></div>
              <div>
                <div className="h-8 bg-gray-200 rounded mb-4"></div>
                <div className="h-6 bg-gray-200 rounded mb-4"></div>
                <div className="h-4 bg-gray-200 rounded mb-2"></div>
                <div className="h-4 bg-gray-200 rounded mb-4"></div>
                <div className="h-10 bg-gray-200 rounded"></div>
              </div>
            </div>
          </div>
        </main>
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="min-h-screen bg-gray-50">
        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center py-12">
            <h1 className="text-2xl font-bold text-gray-900 mb-4">
              {error || 'Product not found'}
            </h1>
            <Link href="/products">
              <Button>Back to Products</Button>
            </Link>
          </div>
        </main>
      </div>
    );
  }

  const images = product.images?.length > 0 ? product.images : ['/placeholder-product.jpg'];

  return (
    <div className="min-h-screen bg-gray-50">
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="mb-8">
          <ol className="flex items-center space-x-2 text-sm text-gray-500">
            <li>
              <Link href="/" className="hover:text-gray-700">Home</Link>
            </li>
            <li>/</li>
            <li>
              <Link href="/products" className="hover:text-gray-700">Products</Link>
            </li>
            <li>/</li>
            <li className="text-gray-900">{product.name}</li>
          </ol>
        </nav>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Product Images */}
          <div>
            <div className="relative h-96 mb-4 bg-white rounded-lg overflow-hidden">
              <Image
                src={images[selectedImageIndex]}
                alt={product.name}
                fill
                className="object-cover"
                sizes="(max-width: 768px) 100vw, 50vw"
              />
              {product.inventory === 0 && (
                <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
                  <span className="text-white font-semibold text-lg">Out of Stock</span>
                </div>
              )}
            </div>
            
            {/* Image Thumbnails */}
            {images.length > 1 && (
              <div className="flex space-x-2 overflow-x-auto">
                {images.map((image, index) => (
                  <button
                    key={index}
                    onClick={() => setSelectedImageIndex(index)}
                    className={`flex-shrink-0 w-16 h-16 rounded-md overflow-hidden border-2 ${
                      selectedImageIndex === index ? 'border-blue-600' : 'border-gray-200'
                    }`}
                  >
                    <Image
                      src={image}
                      alt={`${product.name} ${index + 1}`}
                      width={64}
                      height={64}
                      className="object-cover w-full h-full"
                    />
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Product Info */}
          <div>
            <h1 className="text-3xl font-bold text-gray-900 mb-4">{product.name}</h1>
            
            <div className="flex items-center space-x-4 mb-6">
              <span className="text-3xl font-bold text-gray-900">
                {formatPrice(product.price)}
              </span>
              {product.compareAtPrice && product.compareAtPrice > product.price && (
                <span className="text-xl text-gray-500 line-through">
                  {formatPrice(product.compareAtPrice)}
                </span>
              )}
            </div>

            <div className="mb-6">
              <p className="text-gray-700 leading-relaxed">{product.description}</p>
            </div>

            {/* Stock Status */}
            <div className="mb-6">
              {product.inventory > 0 ? (
                <div className="flex items-center text-green-600">
                  <span className="w-2 h-2 bg-green-600 rounded-full mr-2"></span>
                  <span className="font-medium">
                    {product.inventory > 10 ? 'In Stock' : `Only ${product.inventory} left`}
                  </span>
                </div>
              ) : (
                <div className="flex items-center text-red-600">
                  <span className="w-2 h-2 bg-red-600 rounded-full mr-2"></span>
                  <span className="font-medium">Out of Stock</span>
                </div>
              )}
            </div>

            {/* Quantity and Add to Cart */}
            {product.inventory > 0 && (
              <div className="mb-6">
                <div className="flex items-center space-x-4 mb-4">
                  <label className="text-sm font-medium text-gray-700">Quantity:</label>
                  <div className="flex items-center border border-gray-300 rounded-md">
                    <button
                      onClick={() => setQuantity(Math.max(1, quantity - 1))}
                      className="px-3 py-1 text-gray-600 hover:text-gray-800"
                      disabled={quantity <= 1}
                    >
                      -
                    </button>
                    <span className="px-4 py-1 border-x border-gray-300">{quantity}</span>
                    <button
                      onClick={() => setQuantity(Math.min(product.inventory, quantity + 1))}
                      className="px-3 py-1 text-gray-600 hover:text-gray-800"
                      disabled={quantity >= product.inventory}
                    >
                      +
                    </button>
                  </div>
                </div>

                <Button
                  size="lg"
                  onClick={handleAddToCart}
                  className="w-full sm:w-auto"
                >
                  Add to Cart
                </Button>
              </div>
            )}

            {/* Product Details */}
            <div className="border-t pt-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Product Details</h3>
              <dl className="space-y-2">
                <div className="flex">
                  <dt className="text-sm font-medium text-gray-500 w-24">SKU:</dt>
                  <dd className="text-sm text-gray-900">{product.sku}</dd>
                </div>
                {product.category && (
                  <div className="flex">
                    <dt className="text-sm font-medium text-gray-500 w-24">Category:</dt>
                    <dd className="text-sm text-gray-900">{product.category.name}</dd>
                  </div>
                )}
                {product.specifications && Object.keys(product.specifications).length > 0 && (
                  <>
                    {Object.entries(product.specifications).map(([key, value]) => (
                      <div key={key} className="flex">
                        <dt className="text-sm font-medium text-gray-500 w-24 capitalize">
                          {key.replace(/([A-Z])/g, ' $1').trim()}:
                        </dt>
                        <dd className="text-sm text-gray-900">{String(value)}</dd>
                      </div>
                    ))}
                  </>
                )}
              </dl>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}