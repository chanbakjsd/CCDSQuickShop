<script lang="ts">
	import Icon from '$lib/icon/Icon.svelte'

	interface Props {
		imageURL: string
		name: string
		class: string
	}

	const { imageURL, name, class: klass }: Props = $props()

	let overlayShown = $state(false)
	const updateShown = (newVal: boolean) => () => (overlayShown = newVal)
</script>

<button onclick={updateShown(true)} class={`${klass} cursor-zoom-in`}>
	<img alt={`Picture of ${name}`} src={imageURL} />
</button>

{#if overlayShown}
	<button
		class="fixed bottom-0 left-0 right-0 top-0 z-30 cursor-zoom-out bg-black/50 p-4 lg:p-8 xl:p-16"
		onclick={updateShown(false)}
	>
		<Icon
			name="x"
			class="absolute right-2 top-2 size-12 cursor-pointer rounded-full border border-white bg-brand p-2 text-white"
		/>
		<img class="h-full w-full object-contain" alt={`Picture of ${name}`} src={imageURL} />
	</button>
{/if}
