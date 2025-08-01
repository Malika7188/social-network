import { resolve } from 'path';
import { config } from 'dotenv';



config({ path: resolve('../.env') });

/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    API_URL: process.env.API_URL || "http://localhost:8080/api", 
  },
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: '**', // Accept any hostname
      },
    ],
  },
};

export default nextConfig;
