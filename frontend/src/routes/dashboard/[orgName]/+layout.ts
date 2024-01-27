import type { LayoutLoad } from './$types';

export const prerender = false;

export const load: LayoutLoad = ({ params }) => {
	return params;
};
