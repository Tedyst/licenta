import { webauthnRegisterBegin, webauthnRegisterFinish } from '$lib/client';
import type { Actions } from './$types';

export const actions = {
	start: async ({ fetch }) => {
		const response = await webauthnRegisterBegin(fetch);

		if (!response.success) {
			return { error: 'An error occurred' };
		}

		return {
			setupCredential: response.response
		};
	},
	finish: async ({ request, fetch }) => {
		const data = await request.formData();

		const name = data.get('name')?.toString();
		const d = JSON.parse(data.get('data')?.toString() || '{}') as PublicKeyCredential;

		if (!name || !data) {
			return { error: 'Invalid data' };
		}

		const response = await webauthnRegisterFinish(
			JSON.stringify({
				name,
				...d
			}),
			fetch
		);

		if (!response.success) {
			return { error: 'An error occurred' };
		}

		return {};
	}
} satisfies Actions;
