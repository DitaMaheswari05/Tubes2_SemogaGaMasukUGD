/** @type {import('next').NextConfig} */

console.log('Backend URL:', process.env.BACKEND_URL); 

const nextConfig = {
  async rewrites() {
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080'; // Default to localhost for local development
    return [
      {
        source: '/api/elements',
        destination: `${backendUrl}/api/elements`
      },
      {
        source: '/api/find',
        destination: `${backendUrl}/api/find`
      },
      {
        source: '/api/recipes',
        destination: `${backendUrl}/api/recipes`
      },
      {
        source: '/api/scrape',
        destination: `${backendUrl}/api/scrape`
      },
      {
        source: '/api/svgs/:path*',
        destination: `${backendUrl}/svgs/:path*`
      }
    ];
  }
};

module.exports = nextConfig;