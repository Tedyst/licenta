import { webauthnLoginBegin, webauthnLoginFinish } from '$lib/client';
import { redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent, fetch }) => {
	const parentData = await parent();
	const loginStartData = await webauthnLoginBegin(parentData.username, fetch);

	return { loginStartData, ...parentData };
};

export const actions = {
	default: async ({ request, fetch }) => {
		const data = await request.formData();

		const jsonData = data.get('data')?.toString() || '{}';

		const response = await webauthnLoginFinish(jsonData, fetch);

		if (response.success) {
			redirect(302, '/login/successful');
		}

		console.log(response);
		redirect(302, '/login/webauthn/failed');
	}
} satisfies Actions;
