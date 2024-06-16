import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { logout } from '$lib/client';

export const load: PageServerLoad = async ({ fetch }) => {
	const result = await logout(fetch);

	if (result?.ok) {
		redirect(302, '/login');
	}

	return {};
};
