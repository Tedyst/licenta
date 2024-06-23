import type { HandleFetch } from '@sveltejs/kit';

export const handleFetch = (async ({ event, request, fetch }) => {
	const newRequest = new Request(request, { credentials: 'include' });

	const cookies = event.request.headers.get('cookie');

	newRequest.headers.set('cookie', cookies!);

	return fetch(newRequest).then((response) => {
		const headers = new Headers(response.headers);

		headers.getSetCookie().forEach((setCookie) => {
			const name = setCookie.split('=')[0];
			const value = decodeURI(setCookie).replace(`${name}=`, '').split(';')[0];

			event.cookies.set(name, value, {
				path: '/',
				encode(value: string) {
					return value;
				}
			});
		});

		return response;
	});
}) satisfies HandleFetch;
