/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/elements',
        destination: 'http://localhost:8080/api/elements'
      },
      {
        source: '/api/find',
        destination: 'http://localhost:8080/api/find'
      },
      {
        source: '/api/recipes',
        destination: 'http://localhost:8080/api/recipes'
      },
      {
        source: '/api/svgs/:path*',
        destination: 'http://localhost:8080/svgs/:path*'
      }
    ];
  }
};

module.exports = nextConfig;