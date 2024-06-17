import { validateUsername } from '$lib/login/login';
import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions = {
	password: async ({ request }) => {
		const data = await request.formData();

		const error = validateUsername(data.get('username')?.toString() || null);
		if (error !== null) {
			return { error };
		}

		redirect(302, '/login/password?username=' + data.get('username')?.toString() || '');
	},
	webauthn: async ({ request }) => {
		const data = await request.formData();

		redirect(302, '/login/webauthn?username=' + data.get('username')?.toString() || '');
	}
} satisfies Actions;
