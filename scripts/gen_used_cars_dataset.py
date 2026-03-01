#!/usr/bin/env python3
"""Generate deterministic synthetic used-cars CSV datasets for local demos."""

import argparse
import csv
import random
from datetime import datetime, timedelta
from pathlib import Path


MAKES_AND_MODELS = {
    "Toyota": ["Corolla", "Yaris", "Camry", "RAV4", "Auris"],
    "Volkswagen": ["Golf", "Passat", "Polo", "Tiguan", "Touran"],
    "BMW": ["320", "X1", "X3", "520", "118"],
    "Mercedes-Benz": ["C180", "E220", "A180", "GLA200", "CLA200"],
    "Audi": ["A3", "A4", "A6", "Q3", "Q5"],
    "Ford": ["Focus", "Fiesta", "Mondeo", "Kuga", "Puma"],
    "Skoda": ["Octavia", "Fabia", "Superb", "Kodiaq", "Rapid"],
    "Renault": ["Clio", "Megane", "Captur", "Kadjar", "Talisman"],
    "Hyundai": ["i20", "i30", "Tucson", "Kona", "Elantra"],
    "Kia": ["Ceed", "Sportage", "Rio", "Niro", "Stonic"],
}

FUELS = ["petrol", "diesel", "hybrid", "electric", "lpg"]
TRANSMISSIONS = ["manual", "automatic"]
BODY_TYPES = ["hatchback", "sedan", "suv", "wagon", "coupe"]
COLORS = ["black", "white", "silver", "blue", "red", "gray", "green", "brown"]
CITIES = ["Bucharest", "Cluj-Napoca", "Iasi", "Timisoara", "Constanta", "Brasov"]
SELLER_TYPES = ["dealer", "private"]
CONDITIONS = ["used", "certified_used"]


def generate_row(rng: random.Random, listing_id: int) -> list[str]:
    make = rng.choice(list(MAKES_AND_MODELS.keys()))
    model = rng.choice(MAKES_AND_MODELS[make])
    year = rng.randint(2008, 2025)
    mileage_km = rng.randint(3_000, 320_000)
    fuel = rng.choice(FUELS)
    transmission = rng.choice(TRANSMISSIONS)
    engine_l = round(rng.uniform(1.0, 3.5), 1)
    if fuel == "electric":
        engine_l = 0.0
    body_type = rng.choice(BODY_TYPES)
    color = rng.choice(COLORS)
    city = rng.choice(CITIES)
    base_price = max(1500, int((2026 - year) * -900 + 27_000 + rng.randint(-3500, 4500)))
    if mileage_km > 180_000:
        base_price = int(base_price * 0.8)
    if fuel == "electric":
        base_price = int(base_price * 1.25)
    price_eur = max(1200, base_price)

    reg_month = rng.randint(1, 12)
    reg_day = rng.randint(1, 28)
    first_registration = f"{year:04d}-{reg_month:02d}-{reg_day:02d}"
    seller_type = rng.choice(SELLER_TYPES)
    condition = rng.choice(CONDITIONS)
    listed_at = datetime(2026, 1, 1) + timedelta(days=rng.randint(0, 59), hours=rng.randint(0, 23))

    return [
        f"car-{listing_id:06d}",
        make,
        model,
        str(year),
        str(mileage_km),
        fuel,
        transmission,
        f"{engine_l:.1f}",
        body_type,
        color,
        city,
        str(price_eur),
        first_registration,
        seller_type,
        condition,
        listed_at.isoformat(timespec="seconds") + "Z",
    ]


def write_csv(path: Path, rows: int, seed: int) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    rng = random.Random(seed)

    header = [
        "listing_id",
        "make",
        "model",
        "year",
        "mileage_km",
        "fuel",
        "transmission",
        "engine_l",
        "body_type",
        "color",
        "city",
        "price_eur",
        "first_registration",
        "seller_type",
        "condition",
        "listed_at",
    ]

    with path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(header)
        for i in range(1, rows + 1):
            writer.writerow(generate_row(rng, i))


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate used cars demo CSV.")
    parser.add_argument("--rows", type=int, required=True)
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument("--out", type=Path, required=True)
    args = parser.parse_args()
    write_csv(args.out, args.rows, args.seed)


if __name__ == "__main__":
    main()
