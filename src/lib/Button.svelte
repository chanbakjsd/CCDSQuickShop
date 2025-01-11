<script lang="ts">
	export let disabled = false;
	export let size: 'md' | 'lg' = 'lg';
</script>

<button class={`button-${size}`} on:click {disabled}><slot /></button>

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
</style>
