"use client";

import {
  Button,
  DateField,
  DateRangePicker,
  Label,
  RangeCalendar,
} from "@heroui/react";
import { zodResolver } from "@hookform/resolvers/zod";
import type { DateValue } from "@internationalized/date";
import { getLocalTimeZone, parseDate, today } from "@internationalized/date";
import { Controller, useForm } from "react-hook-form";

import { dateRangeSchema, type DateRangeValues } from "../schemas";
import { FraudTable } from "./fraudTable";

// ── Helpers ────────────────────────────────────────────────────────────────

function todayStr() {
  return today(getLocalTimeZone()).toString();
}

function daysAgoStr(n: number) {
  return today(getLocalTimeZone()).subtract({ days: n }).toString();
}

// ── Presets ────────────────────────────────────────────────────────────────

const PRESETS = [
  { label: "Today", start: () => todayStr(), end: () => todayStr() },
  { label: "Last 7 days", start: () => daysAgoStr(7), end: () => todayStr() },
  { label: "Last 30 days", start: () => daysAgoStr(30), end: () => todayStr() },
  { label: "Last 90 days", start: () => daysAgoStr(90), end: () => todayStr() },
];

// ── View ───────────────────────────────────────────────────────────────────

export function FraudView() {
  const {
    control,
    watch,
    setValue,
    formState: { isValid, errors },
  } = useForm<DateRangeValues>({
    resolver: zodResolver(dateRangeSchema),
    defaultValues: { start: daysAgoStr(7), end: todayStr() },
    mode: "onChange",
  });

  const { start, end } = watch();

  const activePreset =
    PRESETS.find((p) => p.start() === start && p.end() === end)?.label ?? null;

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6 px-4 py-8">
      {/* Date range controls */}
      <div className="flex flex-wrap items-end gap-4 rounded-xl border border-border bg-surface p-4">
        {/* Presets */}
        <div className="flex flex-wrap gap-2 pt-1">
          {PRESETS.map((preset) => (
            <Button
              key={preset.label}
              size="sm"
              variant={activePreset === preset.label ? "primary" : "outline"}
              onPress={() => {
                setValue("start", preset.start(), { shouldValidate: true });
                setValue("end", preset.end(), { shouldValidate: true });
              }}
            >
              {preset.label}
            </Button>
          ))}
        </div>

        {/* DateRangePicker */}
        <div className="ml-auto">
          <Controller
            name="start"
            control={control}
            render={({ field: startField }) => (
              <Controller
                name="end"
                control={control}
                render={({ field: endField }) => {
                  const value: { start: DateValue; end: DateValue } | null =
                    startField.value && endField.value
                      ? {
                          start: parseDate(startField.value),
                          end: parseDate(endField.value),
                        }
                      : null;

                  return (
                    <DateRangePicker
                      value={value}
                      maxValue={today(getLocalTimeZone())}
                      isInvalid={!!errors.end}
                      onChange={(range) => {
                        if (range) {
                          startField.onChange(range.start.toString());
                          endField.onChange(range.end.toString());
                        }
                      }}
                    >
                      <Label>Date Range</Label>
                      <DateField.Group fullWidth>
                        <DateField.Input slot="start">
                          {(segment) => <DateField.Segment segment={segment} />}
                        </DateField.Input>
                        <DateRangePicker.RangeSeparator />
                        <DateField.Input slot="end">
                          {(segment) => <DateField.Segment segment={segment} />}
                        </DateField.Input>
                        <DateField.Suffix>
                          <DateRangePicker.Trigger>
                            <DateRangePicker.TriggerIndicator />
                          </DateRangePicker.Trigger>
                        </DateField.Suffix>
                      </DateField.Group>
                      {errors.end && (
                        <p className="mt-1 text-xs text-danger">
                          {errors.end.message}
                        </p>
                      )}
                      <DateRangePicker.Popover>
                        <RangeCalendar aria-label="Select date range">
                          <RangeCalendar.Header>
                            <RangeCalendar.YearPickerTrigger>
                              <RangeCalendar.YearPickerTriggerHeading />
                              <RangeCalendar.YearPickerTriggerIndicator />
                            </RangeCalendar.YearPickerTrigger>
                            <RangeCalendar.NavButton slot="previous" />
                            <RangeCalendar.NavButton slot="next" />
                          </RangeCalendar.Header>
                          <RangeCalendar.Grid>
                            <RangeCalendar.GridHeader>
                              {(day) => (
                                <RangeCalendar.HeaderCell>
                                  {day}
                                </RangeCalendar.HeaderCell>
                              )}
                            </RangeCalendar.GridHeader>
                            <RangeCalendar.GridBody>
                              {(date) => <RangeCalendar.Cell date={date} />}
                            </RangeCalendar.GridBody>
                          </RangeCalendar.Grid>
                          <RangeCalendar.YearPickerGrid>
                            <RangeCalendar.YearPickerGridBody>
                              {({ year }) => (
                                <RangeCalendar.YearPickerCell year={year} />
                              )}
                            </RangeCalendar.YearPickerGridBody>
                          </RangeCalendar.YearPickerGrid>
                        </RangeCalendar>
                      </DateRangePicker.Popover>
                    </DateRangePicker>
                  );
                }}
              />
            )}
          />
        </div>
      </div>

      {/* Table — only render when date range is valid */}
      {isValid && start && end && (
        <FraudTable
          from={`${start}T00:00:00.000Z`}
          to={`${end}T23:59:59.999Z`}
        />
      )}
    </div>
  );
}
