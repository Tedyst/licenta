<script lang="ts">
	export const passwordError: string | null = null;
</script>

<label class="label" for="password">
	<span class="label-text">Password</span>
</label>
<input
	type="password"
	placeholder="password"
	class="input input-bordered {passwordError
		? 'wiggle input-error'
		: ''} transition-colors duration-300 ease-in-out"
	id="password"
	name="password"
	autocomplete="current-password"
/>
{#if passwordError}
	<div class="label text-error text-xs">
		{passwordError}
	</div>
{/if}

<div class="label">
	<a href="/login/forgot-password" class="label-text-alt link link-hover">Forgot password?</a>
</div>
