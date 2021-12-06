interface RepeatPattern {
    monthFlags: number
    dayFlags: number
    hourFlags: number
    minuteFlags: number
    weekdayFlags: number
}

const Sunday = 0
const Monday = 1
const Tuesday = 2
const Wednesday = 3
const Thursday = 4
const Friday = 5
const Saturday = 6

function createRepeatPattern(months: number[], days: number[], hours: number[], minutes: number[], weekdays: number[]): RepeatPattern {
    let monthFlags = 0
    let dayFlags = 0
    let hourFlags = 0
    let minuteFlags = 0
    let weekdayFlags = 0

    months.forEach(m => monthFlags = monthFlags | (1 << (m - 1)))
    days.forEach(m => dayFlags = dayFlags | (1 << (m - 1)))
    hours.forEach(m => hourFlags = hourFlags | (1 << m))
    minutes.forEach(m => minuteFlags = minuteFlags | (1 << m))
    weekdays.forEach(m => weekdayFlags = weekdayFlags | (1 << m))

    return { monthFlags, dayFlags, hourFlags, minuteFlags, weekdayFlags }
}