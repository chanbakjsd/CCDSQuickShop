import type { Config } from 'tailwindcss';

export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],

	theme: {
		extend: {
			colors: {
				brand: '#0f2b50',
			}
		}
	},

	plugins: []
} satisfies Config;
