/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: "standalone",
  webpack: function (config, _) {
    config.experiments = { asyncWebAssembly: true, layers: true };
    return config;
  },
};

export default nextConfig;
