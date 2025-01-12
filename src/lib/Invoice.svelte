<script lang="ts">
	import { fade } from 'svelte/transition';
	import {
		type CartItem,
		type Coupon,
		type OrderItem,
		applyCoupon,
		calculateCartTotal,
		formatPrice
	} from './cart';
	import InvoiceItem from './InvoiceItem.svelte';
	import { flip } from 'svelte/animate';

	export let items: (CartItem | OrderItem)[];
	export let coupon: Coupon | null;
	export let editable = false;

	$: totalBeforeDiscount = calculateCartTotal(items);
	$: totalAfterDiscount = coupon
		? Math.min(applyCoupon(items, coupon), totalBeforeDiscount)
		: totalBeforeDiscount;
	$: discount = totalBeforeDiscount - totalAfterDiscount;

	const deleteItem = (i: number) => () => {
		items.splice(i, 1);
		items = items;
	};
</script>

<div class="flex flex-col gap-1">
	{#each items as item, i (item)}
		<div transition:fade={{ duration: 100 }} animate:flip={{ duration: 200 }}>
			<InvoiceItem
				bind:item={items[i]}
				{editable}
				deleteItem={editable ? deleteItem(i) : undefined}
			/>
		</div>
	{/each}
</div>
<div class="total">
	{#if discount > 0}
		<span>Subtotal</span>
		<span>S$</span>
		<span>{formatPrice(totalBeforeDiscount / 100)}</span>
		<span>
			Discount
			{#if coupon}({coupon.couponCode}){/if}
		</span>
		<span>- S$</span>
		<span>{formatPrice(discount / 100)}</span>
	{/if}
	<div class="contents text-lg text-black">
		<span>Total</span>
		<span>S$</span>
		<span>{formatPrice(totalAfterDiscount / 100)}</span>
	</div>
</div>

<style lang="postcss">
	.total {
		@apply mt-4 grid gap-x-2 text-sm text-gray-500;
		grid-template-columns: 1fr auto auto;
		span:nth-child(3n + 2) {
			@apply whitespace-nowrap text-right;
		}
		span:nth-child(3n) {
			@apply -ml-1 text-right;
		}
	}
</style>
