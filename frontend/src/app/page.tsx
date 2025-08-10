export default function Home() {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <h1 className="text-2xl font-bold text-gray-900">Ecommerce Store</h1>
            <nav className="flex space-x-8">
              <a href="#" className="text-gray-500 hover:text-gray-900">Products</a>
              <a href="#" className="text-gray-500 hover:text-gray-900">Cart</a>
              <a href="#" className="text-gray-500 hover:text-gray-900">Account</a>
            </nav>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="text-center">
          <h2 className="text-4xl font-bold text-gray-900 mb-4">
            Welcome to Our Store
          </h2>
          <p className="text-xl text-gray-600 mb-8">
            Discover amazing products at great prices
          </p>
          <button className="bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700 transition-colors">
            Shop Now
          </button>
        </div>

        <div className="mt-16">
          <h3 className="text-2xl font-bold text-gray-900 mb-8">Featured Products</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[1, 2, 3].map((item) => (
              <div key={item} className="bg-white rounded-lg shadow-md overflow-hidden">
                <div className="h-48 bg-gray-200"></div>
                <div className="p-6">
                  <h4 className="text-lg font-semibold text-gray-900 mb-2">
                    Product {item}
                  </h4>
                  <p className="text-gray-600 mb-4">
                    Lorem ipsum dolor sit amet, consectetur adipiscing elit.
                  </p>
                  <div className="flex justify-between items-center">
                    <span className="text-2xl font-bold text-gray-900">$99.99</span>
                    <button className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition-colors">
                      Add to Cart
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}