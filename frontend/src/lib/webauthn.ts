import { Base64Binary } from './Base64-binary';

export type PublicKeyCredentialCreationOptionsJSON = {
	challenge: string;
	rp: PublicKeyCredentialRpEntity;
	user: {
		id: string;
		name: string;
		displayName: string;
	};
	pubKeyCredParams: PublicKeyCredentialParameters[];
	timeout: number;
	attestation: AttestationConveyancePreference;
	excludeCredentials?: {
		id: string;
		type: PublicKeyCredentialType;
		transports?: AuthenticatorTransport[];
	}[];
};

export function JSONtoPublicKeyCredentialCreationOptions(
	request: PublicKeyCredentialCreationOptionsJSON
): PublicKeyCredentialCreationOptions {
	return {
		...request,
		challenge: Base64Binary.decode(request.challenge, null),
		user: {
			id: Base64Binary.decode(request.user.id, null),
			name: request.user.name,
			displayName: request.user.displayName
		},
		excludeCredentials: request.excludeCredentials?.map((credential) => {
			return {
				...credential,
				id: Base64Binary.decode(credential.id, null)
			};
		})
	};
}

export type PublicKeyCredentialJSON = {
	id: string;
	type: PublicKeyCredentialType;
	rawId: string;
	response: {
		clientDataJSON: string;
		attestationObject: string;
	};
};

export function PublicKeyCredentialToJSON(credential: any): PublicKeyCredentialJSON {
	return {
		...credential,
		rawId: Base64Binary.encode(new Uint8Array(credential.rawId)),
		response: {
			...credential.response,
			clientDataJSON: Base64Binary.encode(new Uint8Array(credential.response.clientDataJSON)),
			attestationObject: Base64Binary.encode(new Uint8Array(credential.response.attestationObject))
		}
	};
}
