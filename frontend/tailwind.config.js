/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {}
	},
	plugins: [require('daisyui')],
	daisyui: {
		themes: [
			{
				light: {
					primary: '#93f765',
					secondary: '#ffc9ca',
					accent: '#fcaedf',
					neutral: '#191a1f',
					'base-100': '#fcfcfd',
					info: '#21a6ed',
					success: '#76eab8',
					warning: '#997d0a',
					error: '#f32116'
				}
			},
			'dark'
		]
	}
};
