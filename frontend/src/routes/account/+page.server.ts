import type { Actions } from './$types';
import { clientFromFetch } from '$lib/client';

export const actions = {
	default: async ({ request, fetch }) => {
		const data = await request.formData();
		const client = clientFromFetch(fetch);

		const currentPassword = data.get('currentPassword')?.toString() || '';
		const newPassword = data.get('newPassword')?.toString() || '';
		const confirmPassword = data.get('confirmPassword')?.toString() || '';

		if (newPassword !== confirmPassword) {
			return { error: 'Passwords do not match' };
		}

		const response = await client.POST('/users/me/change-password', {
			body: { new_password: newPassword, old_password: currentPassword }
		});
		if (response.data?.success) {
			return { success: true };
		}

		return { error: response.data?.error || 'An error occurred' };
	}
} satisfies Actions;
