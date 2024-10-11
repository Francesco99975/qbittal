import { LitElement, html, css } from "lit";
import { property } from "lit/decorators.js";

export class UCounter extends LitElement {
  @property({ type: String })
  idd: string = "";

  @property({ type: Number })
  min: number = 0;

  @property({ type: Number })
  value: number = 0;

  static styles = css`
    div.flex {
      display: flex;
      align-items: center;
      justify-content: space-evenly;
      width: 100%;
    }

    button {
      border-width: 4px;
      width: 3rem;
      height: 3rem;
      --tw-border-opacity: 1;
      border-color: rgb(var(--color-primary) / var(--tw-border-opacity));
      border-style: solid;
      border-radius: 9999px;
      text-align: center;
      --tw-text-opacity: 1;
      color: rgb(var(--color-primary) / var(--tw-text-opacity));
      font-weight: 700;
      font-size: 1.875rem;
      line-height: 2.25rem;
    }
  `;

  private _increase() {
    this.value = +this.value + 1;
  }

  private _decrease() {
    if (+this.value > +this.min) {
      this.value = +this.value - 1;
    }
  }

  protected render() {
    return html`
      <div class="flex">
        <input
          id="${"qty" + this.idd}"
          min="${this.min.toString()}"
          value="${this.value.toString()}"
          type="hidden"
        />
        <button id="${"dec" + this.idd}" @click="${this._decrease}">-</button>
        <span class="text-xl md:text-3xl font-bold p-2 text-center"
          >${this.value}</span
        >
        <button id="${"inc" + this.idd}" @click="${this._increase}">+</button>
      </div>
    `;
  }
}
