<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from '../$types';

	export let data: PageData;
	export let form: ActionData;
</script>

{#if form?.error}
	<div class="alert alert-error">
		{form.error}
	</div>
{/if}

{#if form?.recoveryCodes}
	<div class="hero min-h-screen bg-base-200">
		<div class="text-center hero-content">
			<div class="max-w-md">
				<h1 class="text-5xl font-bold">Recovery Codes</h1>
				<p class="py-6">
					Here are your recovery codes. Store them in a safe place. You can use these codes to
					regain access to your account if you lose access to your TOTP device.
				</p>
				<ul class="list-disc list-inside">
					{#each form.recoveryCodes as code}
						<li>{code}</li>
					{/each}
				</ul>
			</div>
		</div>
	</div>
{:else}
	<div class="hero min-h-screen bg-base-200">
		<div class="text-center hero-content">
			<div class="max-w-md">
				<h1 class="text-5xl font-bold">Register TOTP</h1>
				<p class="py-6">
					Here is the QR code for setting up Time-based One-Time Password (TOTP) authentication.
					Scan the QR code or use the secret key to set up TOTP authentication.<br />
					After setting up TOTP, you will need to enter the current TOTP token to finish enabling TOTP.
				</p>
				<div class="flex justify-center align-middle w-full flex-col">
					<img src={'data:image/png;base64,' + data?.qrCode} alt="QR Code" />
					{data?.secret}
				</div>
				<form use:enhance method="POST" class="flex gap-2 flex-col mt-4">
					<div class="form-control">
						<label class="label" for="token">
							<span class="label-text">Current Token</span>
						</label>
						<input
							type="text"
							name="token"
							placeholder="Current Token"
							class="input input-bordered"
						/>
					</div>
					<button type="submit" class="btn btn-primary">Enable TOTP</button>
				</form>
			</div>
		</div>
	</div>
{/if}
