<script lang="ts">
	export let disabled = false;
	export let size: 'md' | 'lg' = 'lg';
	export let onClick: () => any;

	let loading = false;
	export const click = async () => {
		loading = true;
		await onClick();
		loading = false;
	};

	$: shouldDisable = disabled || loading;
	$: contentClass = loading ? 'opacity-0' : 'opacity-100';
	$: loaderClass = loading ? 'opacity-100' : 'opacity-0';
</script>

<button class={`relative button-${size}`} on:click={click} disabled={shouldDisable}>
	<div class={`transition-all ${contentClass}`}>
		<slot />
	</div>
	<div class={`absolute bottom-1/3 left-1/3 right-1/3 top-1/3 flex ${loaderClass}`}>
		<div class="loader mx-auto"></div>
	</div>
</button>

<style lang="postcss">
	button {
		@apply rounded-full bg-brand px-6 py-2 text-lg text-white transition-all;
		@apply flex items-center justify-center gap-2;
	}
	button:hover {
		@apply scale-105;
	}
	button:active {
		@apply scale-100;
	}
	button:disabled {
		@apply scale-100 cursor-not-allowed bg-gray-300 text-gray-700;
	}

	button.button-md {
		@apply px-4 py-1 text-base;
	}

	.loader {
		@apply max-h-full max-w-full;
		aspect-ratio: 2;
		--_g: no-repeat radial-gradient(circle closest-side, #777 90%, #0000);
		background:
			var(--_g) 0% 50%,
			var(--_g) 50% 50%,
			var(--_g) 100% 50%;
		background-size: calc(100% / 3) 50%;
		animation: l3 1s infinite linear;
	}
	@keyframes l3 {
		20% {
			background-position:
				0% 0%,
				50% 50%,
				100% 50%;
		}
		40% {
			background-position:
				0% 100%,
				50% 0%,
				100% 50%;
		}
		60% {
			background-position:
				0% 50%,
				50% 100%,
				100% 0%;
		}
		80% {
			background-position:
				0% 50%,
				50% 50%,
				100% 100%;
		}
	}
</style>
