import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { loginTOTP } from '$lib/client';

export const actions = {
	default: async ({ request, fetch }) => {
		const data = await request.formData();

		const code = data.get('token')?.toString() || '';

		const response = await loginTOTP(code, fetch);

		if (response.success) {
			redirect(302, '/login/successful');
		}

		return {
			error: response?.errors?.code?.join(';') || response?.error || 'Internal server error'
		};
	}
} satisfies Actions;
