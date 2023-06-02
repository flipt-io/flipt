declare module 'nightwind/helper' {
  /**
   * To initialize nightwind, add the following script tag to the head element of your pages.
   *
   * ```
   * import nightwind from "nightwind/helper"
   *
   * export default function Layout() {
   *   return (
   *     <>
   *       <Head>
   *         <script dangerouslySetInnerHTML={{ __html: nightwind.init() }} />
   *       </Head>
   *       // ...
   *     </>
   *   )
   * }
   * ```
   */
  declare function init(): string;
  /**
   * Similarly, you can use the toggle function to switch between dark and light mode.
   * ```
   * import nightwind from "nightwind/helper"
   *
   * export default function Navbar() {
   *   return (
   *     // ...
   *     <button onClick={() => nightwind.toggle()}></button>
   *     // ...
   *   )
   * }
   * ```
   */
  declare function toggle(): void;
  /**
   * You can use the enable function to specifically switch between dark and light mode.
   * ```
   * import nightwind from "nightwind/helper"
   *
   * export default function Navbar() {
   *   return (
   *     // ...
   *     <button onClick={() => nightwind.enable(true)}></button>
   *     // ...
   *   )
   * }
   * ```
   */
  declare function enable(dark: boolean): void;
}
