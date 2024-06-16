import type { HandleFetch } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

export const handleFetch = (async ({ event, request, fetch }) => {
	if (!request.url.startsWith(env.PUBLIC_BACKEND_URL)) {
		request = new Request(env.PUBLIC_BACKEND_URL + request.url, {
			...request,
			credentials: 'include'
		});
	}

	const cookies = event.request.headers.get('cookie');

	request.headers.set('cookie', cookies!);

	return fetch(request);
}) satisfies HandleFetch;
