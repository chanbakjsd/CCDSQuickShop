export const formatPrice = (price: number) => price.toFixed(2)

export type CartItem = {
	id: string
	name: string
	variant: {
		type: string
		option: string
	}[]
	imageURL: string
	amount: number
	unitPrice: number // The unit price of the item in cents.
}

export type Coupon = {
	requirements: Requirement[]
	couponCode: string
	discount: Discount
};

export type Requirement = {
	type: "purchase_count"
	amount: number
}

// Discount is used for previewing discounts in the cart.
export type Discount = {
	type: "percentage"
	amount: number // 40% discount (i.e. pay 60%) is represented as 40.
}

export const calculateCartTotal = (cart: CartItem[]) =>
	cart.reduce<number>((total, item) => total + item.unitPrice * item.amount, 0);

export const applyCoupon = (cart: CartItem[], coupon: Coupon) => {
	const cartTotal = calculateCartTotal(cart);
	switch (coupon.discount.type) {
		case 'percentage':
			return Math.ceil((cartTotal * (100 - coupon.discount.amount)) / 100);
	}
};

export const checkRequirement = (cart: CartItem[], requirement: Requirement) => {
	switch (requirement.type) {
		case 'purchase_count':
			return cart.reduce<number>((total, item) => total + item.amount, 0) >= requirement.amount;
	}
};
