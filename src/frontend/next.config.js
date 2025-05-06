// next.config.js
module.exports = {
  async rewrites() {
    return [
      {
        source: "/api/find/:path*",
        destination: "http://localhost:8080/api/find/:path*", // your Go server
      },
      {
        source: "/api/recipes/:path*",
        destination: "http://localhost:8080/api/recipes/:path*",
      },
    ];
  },
};
