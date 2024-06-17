import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = ({ url }) => {
	const username = url.searchParams.get('username') || '';

	return { username };
};
