<script lang="ts">
	export let loading: boolean = false;
</script>

<h1 class="text-2xl font-bold">Login using a Passkey</h1>
<div class="label flex flex-col justify-center items-center">
	<div class="w-max h-max relative">
		<div class="loading loading-ring w-32" />
		<div class="absolute top-[50%] left-[50%] -translate-x-1/2 -translate-y-1/2">
			<i class="material-symbols-outlined text-5xl">passkey</i>
		</div>
	</div>
</div>

<h1 class="text-sm font-bold">
	Please insert your passkey into your device and press the button on it
</h1>

<div class="label mt-8">
	<a href="/login" class="label-text-alt link link-hover">
		Click here to go back to the login page
	</a>
</div>
