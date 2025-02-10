<script lang="ts">
	import type { CartItem } from '$lib/cart';
	import CartIcon from '$lib/CartIcon.svelte';
	import Header from '$lib/Header.svelte';

	import type { PageData } from './$types';
	import Cart from './Cart.svelte';
	import MerchSelection from './MerchSelection.svelte';

	const { data }: { data: PageData } = $props();

	let openCart = $state(false);

	let cart: CartItem[] = $state([]);
	const addToCart = (item: CartItem) => {
		// Try to search for existing entries and add amount instead.
		for (let i = 0; i < cart.length; i++) {
			if (
				cart[i].id !== item.id ||
				cart[i].variant.length !== item.variant.length ||
				cart[i].amount + item.amount > 50
			) {
				continue;
			}
			const variantMatch = cart[i].variant.every((x) =>
				item.variant.some((y) => x.type === y.type && x.option === y.option)
			);
			if (!variantMatch) {
				continue;
			}
			// This is a match! Just add amount.
			cart[i].amount += item.amount;
			return;
		}
		cart = [...cart, item];
	};

	const cartColor = $derived(cart.length > 0 ? 'bg-red-300' : 'bg-gray-300');
	const cartItemCount = $derived(cart.reduce((acc, x) => x.amount + acc, 0));
</script>

<svelte:head><title>SCDS Merch Store</title></svelte:head>

<main class="grid grid-cols-1 md:grid-cols-3">
	<div class="left-panel">
		<Header cls="bg-white">
			<button
				class="absolute right-4 flex justify-center gap-1 md:hidden"
				onclick={() => (openCart = true)}
			>
				<CartIcon />
				<div class={`${cartColor} rounded-full px-2`}>{cartItemCount}</div>
			</button>
		</Header>
		<div class="px-2 md:py-2">
			<MerchSelection items={data.items} addItem={addToCart} />
		</div>
	</div>
	<div class="right-panel" class:openCart>
		<Cart
			bind:cart
			availableCoupons={data.coupons}
			closeClass="md:hidden"
			close={() => (openCart = false)}
		/>
	</div>
</main>

<style lang="postcss">
	.left-panel {
		@apply col-span-2 h-fit border-gray-400 pb-4 pt-2;
		@apply flex flex-col gap-4;
		@apply md:mx-0 md:my-2 md:border-b-0 md:border-r md:px-4 md:py-0;
		@apply md:h-[calc(100vh-1rem)] md:overflow-y-auto;
	}
	.right-panel {
		@apply h-screen overflow-y-auto px-4 py-4;
		@apply fixed right-0 top-0 w-screen bg-white;
		@apply hidden transition-all;
		@apply md:relative md:block md:w-full;
	}
	.right-panel.openCart {
		@apply block;
	}
</style>
