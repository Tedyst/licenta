import type { LayoutLoad } from './$types';

export const load: LayoutLoad = ({ url }) => {
	const username = url.searchParams.get('username') || '';

	return { username };
};
