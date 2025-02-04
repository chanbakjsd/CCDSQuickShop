<script lang="ts">
	import IconX from '$lib/IconX.svelte';

	export let imageURL: string;
	export let name: string;
	export let cls: string;

	let overlay = false;
	const showOverlay = () => (overlay = true);
	const hideOverlay = () => (overlay = false);
</script>

<button on:click={showOverlay} class="cursor-zoom-in">
	<img class={cls} alt={`Picture of ${name}`} src={imageURL} />
</button>

{#if overlay}
	<button
		class="fixed bottom-0 left-0 right-0 top-0 z-30 cursor-zoom-out bg-black/50 p-4 lg:p-8 xl:p-16"
		on:click={hideOverlay}
	>
		<span
			class="absolute right-2 top-2 cursor-pointer rounded-full border border-white bg-brand p-2 text-white"
		>
			<IconX />
		</span>
		<img class="h-full w-full" alt={`Picture of ${name}`} src={imageURL} />
	</button>
{/if}
