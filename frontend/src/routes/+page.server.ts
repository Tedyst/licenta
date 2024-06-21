import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const parentData = await parent();

	if (!parentData.user) {
		redirect(302, '/login');
	} else {
		redirect(302, '/dashboard');
	}

	return parentData;
};
