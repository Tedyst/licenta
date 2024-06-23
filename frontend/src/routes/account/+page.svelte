<script lang="ts">
	import { enhance } from '$app/forms';
	import type { ActionData } from './types';

	export let form: ActionData;
</script>

<div class="hero bg-base-200">
	<div class="hero-content text-center">
		<div class="max-w-md">
			<h1 class="text-5xl font-bold">Account Management</h1>
			<p class="py-6">
				Here you can manage your account settings. You can enable or disable two-factor
				authentication and webauthn, and change your password.
			</p>
			<!-- Password Change Form -->
			<div class="card flex-shrink-0 w-full max-w-sm shadow-2xl bg-base-100">
				<form method="POST" use:enhance class="card-body">
					{#if form?.error}
						<p class="text-red-500">{form.error}</p>
					{/if}
					{#if form?.success}
						<p class="text-green-500">{form.success}</p>
					{/if}
					<div class="form-control">
						<label class="label" for="current-password">
							<span class="label-text">Current Password</span>
						</label>
						<input
							type="password"
							name="current-password"
							placeholder="Current Password"
							class="input input-bordered"
							autoComplete="current-password"
						/>
					</div>
					<div class="form-control">
						<label class="label" for="new-password">
							<span class="label-text">New Password</span>
						</label>
						<input
							type="password"
							name="new-password"
							placeholder="New Password"
							class="input input-bordered"
							autocomplete="new-password"
						/>
					</div>
					<div class="form-control">
						<label class="label" for="confirm-password">
							<span class="label-text">Confirm New Password</span>
						</label>
						<input
							type="password"
							placeholder="Confirm New Password"
							class="input input-bordered"
							name="confirm-password"
							autocomplete="new-password"
						/>
					</div>
					<div class="form-control mt-6">
						<button class="btn btn-primary">Change Password</button>
					</div>
				</form>
			</div>
			<!-- Links to TOTP and Webauthn Settings -->
			<div class="py-6 flex flex-col gap-4 max-w-sm">
				<a href="/account/totp" class="btn btn-primary">Change TOTP Settings</a>
				<a href="/account/webauthn/" class="btn btn-primary">Modify Webauthn Keys</a>
			</div>
		</div>
	</div>
</div>
