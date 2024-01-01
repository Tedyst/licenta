import createClient from 'openapi-fetch';
import type { paths } from './api/v1';
import type { PublicKeyCredentialCreationOptionsJSON } from './webauthn';

let token: string | null = null;

export async function csrfFetch(input: RequestInfo | URL, init?: RequestInit | undefined) {
	if (token == null) {
		token = await getCSRFToken(input);
	}
	return await fetch(input, {
		...init,
		headers: {
			...init?.headers,
			'X-CSRF-Token': token || ''
		}
	});
}

async function getCSRFToken(input: RequestInfo | URL) {
	const optionsResponse = await fetch(input, {
		method: 'OPTIONS',
		headers: {
			'Content-Type': 'application/json'
		}
	});
	return optionsResponse.headers.get('X-CSRF-Token');
}

type webauthnRegisterBeginResponse = {
	success: true;
	response: PublicKeyCredentialCreationOptionsJSON;
};

export async function webauthnRegisterBegin(): Promise<webauthnRegisterBeginResponse> {
	return await csrfFetch('/api/auth/webauthn/register/begin', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		}
	}).then((response) => {
		if (response.ok) {
			return response.json() as Promise<webauthnRegisterBeginResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

type webauthnRegisterFinishResponse = {
	success: true;
};

export async function webauthnRegisterFinish(
	body: string
): Promise<webauthnRegisterFinishResponse> {
	return await csrfFetch('/api/auth/webauthn/register/finish', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body
	}).then((response) => {
		if (response.ok) {
			return response.json() as Promise<webauthnRegisterFinishResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

export async function webauthnLoginBegin() {
	return await csrfFetch('/api/auth/webauthn/login/begin', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		}
	});
}

export async function webauthnLoginFinish(body: string) {
	return await csrfFetch('/api/auth/webauthn/login/finish', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body
	});
}

const client = createClient<paths>({ fetch: csrfFetch, baseUrl: '/api/v1' });

export default client;
