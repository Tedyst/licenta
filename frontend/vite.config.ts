import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit()],
	test: {
		include: ['src/**/*.{test,spec}.{js,ts}']
	},
	server: {
		host: true,
		proxy: {
			'/api/': {
				target: process.env.PUBLIC_BACKEND_URL
					? process.env.PUBLIC_BACKEND_URL
					: 'http://localhost:5000',
				changeOrigin: true,
				preserveHeaderKeyCase: true
			}
		}
	}
});
