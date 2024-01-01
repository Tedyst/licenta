import { writable } from 'svelte/store';

export function validatePassword(password: string): string | null {
	return password.length >= 8 ? null : 'Password must be at least 8 characters long';
}

export function validateUsername(username: string): string | null {
	if (username.length < 3) {
		return 'Username must be at least 3 characters long';
	}
	if (!/^[a-zA-Z0-9]+$/.test(username)) {
		return 'Username can only contain letters and numbers';
	}
	return null;
}

export function validateTOTPToken(token: string): string | null {
	if (!/^[0-9]{6}$/.test(token)) {
		return 'Token must be 6 digits long';
	}
	return null;
}

export const username = writable('');
