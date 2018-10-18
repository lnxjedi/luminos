```yaml
#luminos
# This is an example of Luminos frontmatter.
# Generate a TOC from markdown and set .TOC flag
MDTOC: true
```
# It works!

This is **Luminos**[^lum] a markdown server written in **Go**[^go]. Originally written[^orig] by **José Nieto**[^jose], functionality has been extended and modified (with e.g. search and modular templates) to deliver more of a lightweight CMS by **David Parsley**[^parse].

**Luminos**[^lum] is an Open Source project, feel free to browse and hack the
source[^lum].

Thanks for using **Luminos**!

# A few markdown examples

[Markdown](http://daringfireball.net/projects/markdown/) is a very comfortable
format for writing documents in plain text format.

Here are some examples on how your markdown code would be translated into HTML
by [Luminos][3].

## Emphasis
<table class="table">
  <thead>
    <tr>
      <th>Markdown code</th>
      <th>Result</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
        <code>**Bold text**</code>
      </td>
      <td>
        <strong>Bold text</strong>
      </td>
    </tr>
    <tr>
      <td>
        <code>*Italics*</code>
      </td>
      <td>
        <em>Italics</em>
      </td>
    </tr>
    <tr>
      <td>
        <code>~~Striked-through~~</code>
      </td>
      <td>
        <del>Striked-through</del>
      </td>
    </tr>
  </tbody>
</table>

## Headings
<table class="table">
  <thead>
    <tr>
      <th>Markdown code</th>
      <th>Result</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
        <code># First level header</code>
      </td>
      <td>
        <h1>First level header</h1>
      </td>
    </tr>
    <tr>
      <td>
        <code>## Second level header</code>
      </td>
      <td>
        <h2>Second level header</h2>
      </td>
    </tr>
    <tr>
      <td>
        <code>### Third level header</code>
      </td>
      <td>
        <h3>Third level header</h3>
      </td>
    </tr>
    <tr>
      <td>
        <code>#### Fourth level header</code>
      </td>
      <td>
        <h4>Fourth level header</h4>
      </td>
    </tr>
    <tr>
      <td>
        <code>##### Fifth level header</code>
      </td>
      <td>
        <h5>Fifth level header</h5>
      </td>
    </tr>
  </body>
</table>

## Links
<table class="table">
  <thead>
    <tr>
      <th>Markdown code</th>
      <th>Result</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
        <code>[The Go Programming Language](http://golang.org)</code>
      </td>
      <td>
        <a href="http://golang.org">The Go Programming Language</a>
      </td>
    </tr>
  </body>
</table>

## Lists and Tables
<table class="table">
  <thead>
    <tr>
      <th>Markdown code</th>
      <th>Result</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
<pre><code>* List item 1
* List item 2
* List item 3</code></pre>
      </td>
      <td>
        <ul>
          <li>List item 1</li>
          <li>List item 2</li>
          <li>List item 3</li>
        </ul>
      </td>
    </tr>
    <tr>
      <td>
<pre><code>1. List item 1
2. List item 2
3. List item 3</code></pre>
      </td>
      <td>
        <ol>
          <li>List item 1</li>
          <li>List item 2</li>
          <li>List item 3</li>
        </ol>
      </td>
    </tr>
    <tr>
      <td>
<pre><code>Name    | Age
--------|------
Bob     | 27
Alice   | 23</code></pre>
      </td>
      <td>
        <table>
          <thead>
            <tr>
              <td>Name</td>
              <td>Age</td>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>Bob</td>
              <td>27</td>
            </tr>
            <tr>
              <td>Alice</td>
              <td>23</td>
            </tr>
          </tbody>
        </table>
      </td>
    </tr>
  </body>
</table>

## Code

### Code string formatting

Using:

<pre>Example of `code formatted strings`.</pre>

Gives:

Example of `code formatted strings`.

### Code fences

Using:

<pre>```python
import sys

sys.path.append("/usr/local/python")
```
</pre>

Gives:

```python
import sys

sys.path.append("/usr/local/python")
```

[^lum]: http://golang.org
[^go]: https://github.com/lnxjedi/luminos
[^jose]: https://github.com/xiam
[^orig]: https://github.com/xiam/luminos
[^parse]: https://github.com/parsley42
