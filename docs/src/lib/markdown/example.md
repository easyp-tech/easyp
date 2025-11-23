---
title: "Example Markdown Document"
description: "A comprehensive example showcasing all markdown features"
toc: true
---

# Example Markdown Document

[[toc]]

This is an example markdown document that demonstrates all the features supported by the EasyP Markdown Library.

<div class="tip custom-block" style="padding-top: 8px">

This is a **tip block** that shows how HTML content is rendered within markdown. You can include any HTML elements here!

</div>

## Table of Contents Features

The table of contents is automatically generated from headings in your document. You can control which headings are included by adjusting the `tocDepth` configuration.

### Nested Heading Example

This is a third-level heading that will appear nested in the TOC.

#### Fourth Level Heading

And this is a fourth-level heading for even deeper nesting.

## Code Examples with Focus

Here's a JavaScript example with focus highlighting:

```javascript
function greet(name) {
    console.log("Setting up greeting..."); 
    const message = `Hello, ${name}!`; // [!code focus]
    console.log("Greeting prepared");
    return message; // [!code focus]
}

// Usage example
const greeting = greet("World"); // [!code focus]
console.log(greeting);
```

And here's a Python example:

```python
def calculate_sum(numbers):
    total = 0
    for num in numbers: # [!code focus]
        total += num # [!code focus]
    return total

# Example usage
result = calculate_sum([1, 2, 3, 4, 5]) # [!code focus]
print(f"Sum: {result}")
```

## Custom Blocks

### Warning Block

<div class="warning custom-block">

**Warning**: This is a warning block. Use it to highlight important information that users should be aware of.

</div>

### Danger Block

<div class="danger custom-block">

**Danger**: This indicates something potentially harmful or destructive. Use sparingly and only when absolutely necessary.

</div>

### Details Block

<div class="details custom-block">

**Click to expand**: This is a collapsible details block that can contain additional information.

You can include multiple paragraphs, lists, and even code blocks inside details blocks.

</div>

## Links and Navigation

Here are examples of different types of links:

- [Internal link to another page](./what-is.md)
- [External link to GitHub](https://github.com/easyp-tech/easyp)
- [Anchor link to code section](#code-examples-with-focus)
- [Link to a specific heading](#nested-heading-example)

## Lists and Formatting

### Unordered Lists

- First item
- Second item with **bold text**
- Third item with *italic text*
- Fourth item with `inline code`

### Ordered Lists

1. First numbered item
2. Second numbered item
3. Third numbered item with a [link](https://example.com)

### Task Lists

- [x] Completed task
- [ ] Incomplete task
- [x] Another completed task

## Tables

| Feature | Supported | Notes |
|---------|-----------|-------|
| TOC Generation | ✅ | Automatic from headings |
| Code Focus | ✅ | Using `[!code focus]` |
| HTML Blocks | ✅ | Custom tip, warning, danger blocks |
| Link Processing | ✅ | Internal and external links |

## Blockquotes

> This is a blockquote. It can be used to highlight important information or quotes from other sources.
> 
> Blockquotes can span multiple lines and paragraphs.

## Inline Code and Formatting

Here's some text with `inline code`, **bold text**, *italic text*, and ~~strikethrough text~~.

You can also use combinations like ***bold and italic*** or `**bold code**`.

## Mathematical Expressions

While mathematical expressions aren't directly supported in the base library, you can always extend it with additional plugins for LaTeX support.

## Images

![Example Image](https://via.placeholder.com/600x300.png?text=Example+Image)

## Horizontal Rules

Use horizontal rules to separate sections:

---

## Conclusion

This example demonstrates the comprehensive markdown processing capabilities of the EasyP Markdown Library. The library handles:

1. **Table of Contents**: Automatic generation from headings
2. **Code Focus**: Highlighting specific lines in code blocks
3. **Custom HTML Blocks**: Tip, warning, danger, and details blocks
4. **Link Processing**: Internal navigation and external links
5. **Standard Markdown**: All common markdown features

For more information, visit the [EasyP documentation](https://github.com/easyp-tech/easyp).