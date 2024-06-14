import { validateEmail, validatePassword, validateUsername } from '$lib/login/login';
import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { register } from '$lib/client';

export const actions = {
	default: async ({ request, fetch }) => {
		const data = await request.formData();

		const username = data.get('username')?.toString() || '';
		const email = data.get('email')?.toString() || '';
		const password = data.get('password')?.toString() || '';

		const response = await register({ username, email, password }, fetch);

		if (response === undefined) {
			return { error: 'Failed to fetch' };
		}

		if (response.success) {
			redirect(302, '/login');
		}

		return {
			usernameError: validateUsername(username),
			emailError: validateEmail(email),
			passwordError: validatePassword(password),
			error: response.message || null
		};
	}
} satisfies Actions;
