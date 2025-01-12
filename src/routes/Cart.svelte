<script lang="ts">
	import { checkout, fetchCoupon } from '$lib/api';
	import { type CartItem, type Coupon, applyCoupon, checkRequirement } from '$lib/cart';
	import Button from '$lib/Button.svelte';
	import Input from '$lib/Input.svelte';
	import Invoice from '$lib/Invoice.svelte';

	export let cart: CartItem[];
	export let availableCoupons: Coupon[];

	const validateName = (name: string) => name !== '';
	const validateMatricNum = (matricNum: string) => /^[UG]\d{7}[A-Z]$/.test(matricNum);
	const validateEmail = (email: string) => /^[A-Za-z0-9]+$/.test(email);

	const findBestCoupon = (cart: CartItem[], coupons: Coupon[]) => {
		return coupons.reduce<Coupon | null>((prev, candidate) => {
			// Choose the better coupon (prev).
			if (prev && applyCoupon(cart, prev) < applyCoupon(cart, candidate)) return prev;
			// Use the candidate if we can.
			if (candidate.requirements.every((x) => checkRequirement(cart, x))) return candidate;
			// Otherwise, use whatever we had previously.
			return prev;
		}, null);
	};
	const searchCoupon = async (cart: CartItem[], couponCode: string) => {
		let candidate = availableCoupons.find((x) => x.couponCode === couponCode);
		if (!candidate) {
			candidate = await fetchCoupon(couponCode);
		}
		if (candidate.requirements.every((x) => checkRequirement(cart, x))) return candidate;
		throw new Error('Coupon requirement not matched');
	};

	$: bestCoupon = findBestCoupon(cart, availableCoupons);

	let userName = '';
	let userMatricNumber = '';
	let userEmail = '';
	$: couponCode = bestCoupon?.couponCode || '';
	let coupon: Coupon | undefined = undefined;
	let couponInvalid = false;

	$: {
		if (couponCode) {
			couponInvalid = false;
			searchCoupon(cart, couponCode)
				.then((x) => {
					couponInvalid = false;
					coupon = x;
				})
				.catch(() => {
					couponInvalid = true;
				});
		}
	}

	$: checkoutValid =
		cart.length > 0 &&
		validateName(userName) &&
		validateMatricNum(userMatricNumber) &&
		validateEmail(userEmail) &&
		(couponCode === '' || coupon);

	const processCheckout = async () => {
		const checkoutURL = await checkout(
			cart,
			userName,
			userMatricNumber,
			userEmail + '@e.ntu.edu.sg',
			bestCoupon?.couponCode
		);
		window.location.href = checkoutURL;
	};

	let checkoutButton: Button;
	const clickCheckout = () => checkoutButton.click();
</script>

<div class="flex min-h-full flex-col justify-between gap-4">
	<div>
		<h1 class="text-2xl">Cart</h1>
		<div class="full flex flex-col gap-2 lg:px-8">
			<div class="my-2">
				{#if cart.length === 0}
					<p class="text-center text-sm italic">The cart is currently empty.</p>
				{:else}
					<Invoice bind:items={cart} coupon={bestCoupon} editable />
				{/if}
			</div>
		</div>
	</div>
	<div class="flex flex-col gap-2">
		<form class="grid grid-cols-1 xl:grid-cols-2" on:submit={clickCheckout}>
			<Input label="Name" bind:value={userName} validate={validateName} />
			<Input label="Matric Number" bind:value={userMatricNumber} validate={validateMatricNum} />
			<div class="flex items-end xl:col-span-2">
				<div class="min-w-0 flex-grow">
					<Input label="Email" bind:value={userEmail} validate={validateEmail} />
				</div>
				<span class="text-lg">@e.ntu.edu.sg</span>
			</div>
			<div class="xl:col-span-2">
				<Input label="Coupon Code" bind:value={couponCode} invalid={couponInvalid} />
			</div>
			<!-- https://stackoverflow.com/questions/4196681/form-not-submitting-when-pressing-enter -->
			<input type="submit" class="hidden" />
		</form>
		<Button disabled={!checkoutValid} onClick={processCheckout} bind:this={checkoutButton}>
			Checkout
		</Button>
	</div>
</div>
