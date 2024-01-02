<div class="hero min-h-screen">
	<div class="hero-content flex-col lg:flex-row-reverse">
		<div class="text-center lg:text-left">
			<h1 class="text-5xl font-bold">Register now!</h1>
			<p class="py-6 md:w-max max-w-md">
				Please appreciate my login animation. I spent 8 hours on it instead of working for my job.
				Help
			</p>
		</div>
		<div
			class="card flex-shrink-0 w-full max-w-sm shadow-2xl bg-base-100 overflow-hidden h-[400px]"
			id="/"
		>
			<div class="card-body absolute justify-center m-0 left-0 right-0 top-0 bottom-0 text-center">
				<slot />
			</div>
		</div>
	</div>
</div>

<svelte:head>
	<title>Register</title>
	<meta name="description" content="Register your account" />
</svelte:head>
