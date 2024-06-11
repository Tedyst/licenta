import createClient, { type MiddlewareRequest } from 'openapi-fetch';
import type { paths } from './api/v1';
import type {
	PublicKeyCredentialCreationOptionsJSON,
	PublicKeyCredentialRequestOptionsJSON
} from './webauthn';
import { env } from '$env/dynamic/public';

export async function csrfFetch(
	input: RequestInfo | URL,
	init?: RequestInit | undefined,
	f: typeof fetch = fetch
) {
	const token = await getCSRFToken(input, f);
	return await f(input, {
		...init,
		headers: {
			...init?.headers,
			'X-CSRF-Token': token || ''
		}
	});
}

async function getCSRFToken(input: RequestInfo | URL, f: typeof fetch = fetch) {
	const optionsResponse = await f(input, {
		method: 'OPTIONS',
		headers: {
			'Content-Type': 'application/json'
		},
		credentials: 'include'
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

type webauthnLoginBeginResponse = {
	success: true;
	response: PublicKeyCredentialRequestOptionsJSON;
};

export async function webauthnLoginBegin(
	username: string | null = null
): Promise<webauthnLoginBeginResponse> {
	return await csrfFetch('/api/auth/webauthn/login/begin', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ username: username })
	}).then((response) => {
		if (response.ok) {
			return response.json() as Promise<webauthnLoginBeginResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

type webauthnLoginFinishResponse = {
	success: true;
};

export async function webauthnLoginFinish(body: string): Promise<webauthnLoginFinishResponse> {
	return await csrfFetch('/api/auth/webauthn/login/finish', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body
	}).then((response) => {
		if (response.ok) {
			return response.json() as Promise<webauthnLoginFinishResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

type LoginResponse = {
	success: boolean;
	totp?: boolean;
	webauthn?: boolean;
	error?: string;
};

export async function login(
	username: string,
	password: string,
	remember: boolean,
	f: typeof fetch = fetch,
	baseURL: string = env.PUBLIC_BACKEND_URL
): Promise<LoginResponse> {
	return await csrfFetch(
		baseURL + '/api/auth/login',
		{
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ username, password, rm: String(remember) })
		},
		f
	).then((response) => {
		if (response?.ok) {
			return response.json() as Promise<LoginResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

type RegisterTOTPBeginResponse = {
	success: true;
};

export async function registerTOTPBegin(): Promise<RegisterTOTPBeginResponse> {
	return await csrfFetch('/api/auth/2fa/totp/setup', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		}
	}).then((response) => {
		if (response?.ok) {
			return response.json() as Promise<RegisterTOTPBeginResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

type RegisterTOTPGetSecretResponse = {
	success: true;
	totp_secret: string;
};

export async function registerTOTPGetSecret(): Promise<RegisterTOTPGetSecretResponse> {
	return await csrfFetch('/api/auth/2fa/totp/confirm', {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json'
		}
	}).then((response) => {
		if (response?.ok) {
			return response.json() as Promise<RegisterTOTPGetSecretResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

type RegisterTOTPFinishResponse = {
	success: true;
	recovery_codes: string[];
};

export async function registerTOTPFinish(code: string): Promise<RegisterTOTPFinishResponse> {
	return await csrfFetch('/api/auth/2fa/totp/confirm', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ code })
	}).then((response) => {
		if (response?.ok) {
			return response.json() as Promise<RegisterTOTPFinishResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

export async function logout(): Promise<void> {
	return await csrfFetch('/api/auth/logout', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		}
	}).then((response) => {
		if (response.ok) {
			return;
		}
		throw new Error('Failed to fetch');
	});
}

type LoginTOTPResponse = {
	success: boolean;
	errors?: {
		code?: string[];
	};
	message?: string;
};

export async function loginTOTP(code: string): Promise<LoginTOTPResponse> {
	return await csrfFetch('/api/auth/2fa/totp/validate', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ code })
	}).then((response) => {
		if (response.ok) {
			return response.json() as Promise<LoginTOTPResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

export type RequestResetPasswordResponse = {
	success: boolean;
};

export async function requestResetPassword(
	username: string,
	f: typeof fetch = fetch,
	baseURL: string = env.PUBLIC_BACKEND_URL
): Promise<RequestResetPasswordResponse> {
	return await csrfFetch(
		baseURL + '/api/auth/recover',
		{
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ username })
		},
		f
	).then((response) => {
		if (response.ok) {
			return response.json() as Promise<RequestResetPasswordResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

export type RegisterUserResponse = {
	success: boolean;
	errors?: {
		username?: string[];
		email?: string[];
		password?: string[];
	};
	message?: string;
};

export type RegisterUserRequest = {
	username: string;
	email: string;
	password: string;
};

export async function register(
	{ username, email, password }: RegisterUserRequest,
	f: typeof fetch = fetch,
	baseURL: string = env.PUBLIC_BACKEND_URL
) {
	return await csrfFetch(
		baseURL + '/api/auth/register',
		{
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ username, email, password, confirm_password: password })
		},
		f
	).then((response) => {
		if (response?.ok) {
			return response.json() as Promise<RegisterUserResponse>;
		}
		throw new Error('Failed to fetch');
	});
}

export const clientFromFetch = (fetch: typeof csrfFetch, origin: string) => {
	const client = createClient<paths>({
		baseUrl: origin + '/api/v1',
		fetch
	});
	const csrfMiddleware = {
		async onRequest(req: MiddlewareRequest) {
			const token = (await getCSRFToken(req.url, fetch)) || 'null';

			req.headers.set('X-CSRF-Token', token);

			return req;
		},
		async onResponse(res: Response) {
			return res;
		}
	};
	client.use(csrfMiddleware);
	return client;
};

export const client = clientFromFetch(fetch, env.PUBLIC_BACKEND_URL);

export default client;
