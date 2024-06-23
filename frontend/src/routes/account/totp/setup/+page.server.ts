import { getQRCode, registerTOTPGetSecret } from '$lib/client';
import type { PageServerLoad, Actions } from './$types';
import { registerTOTPFinish } from '$lib/client';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ fetch, parent, isDataRequest }) => {
	const parentData = await parent();

	if (parentData.user.has_totp && isDataRequest) {
		return {};
	} else if (!isDataRequest && parentData.user.has_totp) {
		redirect(302, '/account/totp');
	}

	const secretResponse = await registerTOTPGetSecret(fetch);

	if (!secretResponse.success) {
		return {
			error: 'An error occurred'
		};
	}

	const qrCode = await getQRCode(fetch);

	return {
		secret: secretResponse.totp_secret,
		qrCode: await qrCode.arrayBuffer().then((buffer) => Buffer.from(buffer).toString('base64'))
	};
};

export const actions = {
	default: async ({ request, fetch }) => {
		const data = await request.formData();

		const token = data.get('token')?.toString() || '';

		if (!token) {
			return { error: 'Invalid token' };
		}

		const response = await registerTOTPFinish(token, fetch);

		if (!response?.success) {
			return { error: 'An error occurred' };
		}

		return {
			recoveryCodes: response.recovery_codes
		};
	}
} satisfies Actions;
