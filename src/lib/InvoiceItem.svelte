<script lang="ts">
	import { type CartItem, type OrderItem, formatPrice } from '$lib/cart'
	import Icon from '$lib/icon/Icon.svelte'
	import ZoomableImage from './ZoomableImage.svelte'

	interface Props {
		item: CartItem | OrderItem
		editable: boolean
		deleteItem?: () => any
	}
	const { item = $bindable(), editable, deleteItem }: Props = $props()
	const itemVariant = $derived(
		typeof item.variant === 'string' ? item.variant : item.variant.map((x) => x.option).join(', ')
	)
</script>

<div class="flex w-full items-center gap-2">
	<ZoomableImage imageURL={item.imageURL} name={item.name} class="size-16" />
	<div class="flex w-full min-w-0 flex-col break-words">
		<p>{item.name}</p>
		<p class="text-xs italic text-gray-500">{itemVariant}</p>
		<div class="flex w-full flex-row items-start justify-between md:flex-col lg:flex-row">
			<div class="flex gap-1">
				{#if editable}
					<select bind:value={item.amount}>
						{#each Array(50) as _, i}
							<option value={i + 1}>{i + 1}</option>
						{/each}
					</select>
				{:else}
					<span>x {item.amount}</span>
				{/if}
				{#if deleteItem}
					<button onclick={deleteItem}>
						<Icon name="trash" class="size-4 text-gray-500" />
					</button>
				{/if}
			</div>
			<p>S$ {formatPrice((item.amount * item.unitPrice) / 100)}</p>
		</div>
	</div>
</div>

<style lang="postcss">
	select {
		@apply rounded-xl p-1 text-center;
	}
</style>
