import type { PageLoad } from './$types';

export const load: PageLoad = ({ url }) => {
	const username = url.searchParams.get('username') || '';

	return { username };
};
