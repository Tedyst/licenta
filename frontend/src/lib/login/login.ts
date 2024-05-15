export function validatePassword(password: string | null): string | null {
	if (password === null) {
		return 'Password cannot be empty';
	}
	if (password.length < 8) {
		return 'Password must be at least 8 characters long';
	}
	if (!/[0-9]/.test(password)) {
		return 'Password must contain at least one number';
	}
	if (!/[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]/.test(password)) {
		return 'Password must contain at least one symbol';
	}
	if (!/[A-Z]/.test(password)) {
		return 'Password must contain at least one uppercase letter';
	}
	if (!/[a-z]/.test(password)) {
		return 'Password must contain at least one lowercase letter';
	}
	return null;
}

export function validateUsername(username: string | null): string | null {
	if (username === null || username.length < 3) {
		return 'Username must be at least 3 characters long';
	}
	if (!/^[a-zA-Z0-9]+$/.test(username)) {
		return 'Username can only contain letters and numbers';
	}
	return null;
}

export function validateEmail(email: string): string | null {
	if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
		return 'Invalid email address';
	}
	return null;
}

export function validateTOTPToken(token: string): string | null {
	if (!/^[0-9]{6}$/.test(token)) {
		return 'Token must be 6 digits long';
	}
	return null;
}
