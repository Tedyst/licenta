import type { Actions } from './$types';
import { disableTOTP, registerTOTPBegin, registerTOTPFinish } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const actions = {
	start: async ({ fetch }) => {
		const beginResponse = await registerTOTPBegin(fetch);

		if (!beginResponse.success) {
			return {
				error: 'An error occurred'
			};
		}

		redirect(302, '/account/totp/setup');
	},
	add: async ({ request, fetch }) => {
		const data = await request.formData();

		const token = data.get('token')?.toString() || '';

		if (!token) {
			return { error: 'Invalid token' };
		}

		const response = await registerTOTPFinish(token, fetch);

		if (!response.success) {
			return { error: 'An error occurred' };
		}

		return {
			recoveryCodes: response.recovery_codes
		};
	},
	remove: async ({ request, fetch }) => {
		const data = await request.formData();

		const token = data.get('token')?.toString() || '';

		if (!token) {
			return { error: 'Invalid token' };
		}

		const response = await disableTOTP(token, fetch);

		if (!response.success) {
			return { error: response?.errors?.code?.[0] || 'An error occurred' };
		}

		return {};
	}
} satisfies Actions;
