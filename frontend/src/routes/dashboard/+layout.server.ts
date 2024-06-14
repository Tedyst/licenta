import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ parent }) => {
	const parentData = await parent();
	if (parentData.user == null) {
		redirect(302, '/login');
	}
	return parentData;
};
