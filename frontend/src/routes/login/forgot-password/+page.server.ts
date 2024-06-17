import { validateUsername } from '$lib/login/login';
import type { Actions } from './$types';
import { requestResetPassword } from '$lib/client';

export const actions = {
	default: async ({ request, fetch }) => {
		const data = await request.formData();

		const username = data.get('username')?.toString() || '';

		const error = validateUsername(username);
		if (error !== null) {
			return { error };
		}

		const response = await requestResetPassword(username, fetch).catch(() => {});

		if (response === undefined) {
			return { error: 'Failed to fetch' };
		}

		return {
			message:
				'An email has been sent to your email address with instructions on how to reset your password.'
		};
	}
} satisfies Actions;
