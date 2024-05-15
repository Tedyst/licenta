import { validatePassword, validateUsername } from '$lib/login/login';
import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { login } from '$lib/client';

export const actions = {
	default: async ({ request, fetch, url }) => {
		const data = await request.formData();

		const username = data.get('username')?.toString() || '';
		const password = data.get('password')?.toString() || '';
		const remember = data.get('remember')?.toString() || '';

		const error = validateUsername(username) || validatePassword(password);
		if (error !== null) {
			return { error };
		}

		const response = await login(username, password, remember === 'on', fetch, url.origin).catch(
			() => {}
		);

		if (response === undefined) {
			return { error: 'Failed to fetch' };
		}

		if (response.success) {
			redirect(302, '/login/successful');
		} else if (response.totp && response.webauthn) {
			redirect(302, '/login/2fa');
		} else if (response.totp) {
			redirect(302, '/login/totp');
		} else if (response.webauthn) {
			redirect(302, '/login/webauthn');
		}
		return { error: response.message || null };
	}
} satisfies Actions;
